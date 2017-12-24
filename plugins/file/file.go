package file

import (
	"fmt"
	"github.com/hpcloud/tail"
	"gopkg.in/tomb.v1" // need tomb.ErrStillAlive
	"os"
	"strings"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges []string) ([]*types.Bridge, error) {

	brs := []*types.Bridge{}

	for i, br := range bridges {
		pair := strings.SplitN(br, ":", 2)
		if len(pair) == 1 {
			pair = append(pair, pair[0])
		}
		config := tail.Config{
			Follow: true,
			ReOpen: true,
			Logger: side.Log,
		}
		t, err := tail.TailFile(pair[0], config)
		if err != nil {
			return brs, err
		}
		b := types.Bridge{ID: i, Channel: pair[1], Pipe: make(chan string)}
		side.Log.Printf("Bridge %d: producer connect to file %s", b.ID, pair[0])
		go reader(side, t, b)
		side.WG.Add(1)
		brs = append(brs, &b)
	}
	return brs, nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges []*types.Bridge) error {

	for _, b := range bridges {
		if b.Channel == "-" {
			// send to STDOUT
			side.Log.Printf("Bridge %d consumer: connect to stdout", b.ID)
			go printer(side, *b)
			side.WG.Add(1)
		} else {
			f, err := os.Create(b.Channel)
			if err != nil {
				return err
			}
			side.Log.Printf("Bridge %d consumer: connect to file %s", b.ID, b.Channel)
			go writer(side, f, *b)
			side.WG.Add(1)
		}
	}
	return nil
}

func reader(side *types.Side, tf *tail.Tail, br types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-tf.Lines:
			if tf.Err().Error() != tomb.ErrStillAlive.Error() {
				side.Log.Printf("warn: Bridge %d side 'in' error: %v", br.ID, tf.Err().Error())
				tf.Stop()
				side.Abort <- br.ID
				return
			}
			br.Pipe <- line.Text
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d producer closed", br.ID)
			tf.Stop() //Cleanup()
			return
		}
	}
}

func writer(side *types.Side, f *os.File, br types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			_, err := f.WriteString(line + "\n")
			if err != nil {
				side.Log.Printf("warn: Bridge %d consumer error: %v", br.ID, err.Error())
				f.Close()
				side.Abort <- br.ID
				return
			}
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d consumer closed", br.ID)
			f.Close()
			return
		}
	}
}

func printer(side *types.Side, br types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			fmt.Println(line)
			// check err
		case <-side.Quit:
			side.Log.Printf("debug: Channel %d consumer closed", br.ID)
			return
		}
	}
}
