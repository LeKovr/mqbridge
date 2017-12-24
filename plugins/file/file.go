package file

import (
	"fmt"
	"github.com/hpcloud/tail"
	"gopkg.in/tomb.v1" // need tomb.ErrStillAlive
	"os"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges types.Bridges) error {
	config := tail.Config{
		Follow: true,
		ReOpen: true,
		Logger: side.Log,
	}

	for _, br := range bridges {
		t, err := tail.TailFile(br.In, config)
		if err != nil {
			return err
		}
		side.Log.Printf("Bridge %d: producer connect to file %s", br.ID, br.In)
		go reader(side, t, br)
		side.WG.Add(1)
	}
	return nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges types.Bridges) error {

	for _, br := range bridges {
		if br.Out == "-" {
			// send to STDOUT
			side.Log.Printf("Bridge %d consumer: connect to stdout", br.ID)
			go printer(side, *br)
			side.WG.Add(1)
		} else {
			f, err := os.Create(br.Out)
			if err != nil {
				return err
			}
			side.Log.Printf("Bridge %d consumer: connect to file %s", br.ID, br.Out)
			go writer(side, f, br)
			side.WG.Add(1)
		}
	}
	return nil
}

func reader(side *types.Side, tf *tail.Tail, br *types.Bridge) {
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
			side.Log.Println("debug: BRIN ", br.ID, " ", line.Text)
			br.Pipe <- line.Text
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d producer closed", br.ID)
			tf.Stop() //Cleanup()
			return
		}
	}
}

func writer(side *types.Side, f *os.File, br *types.Bridge) {
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
