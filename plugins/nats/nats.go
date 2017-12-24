package nats

import (
	"strings"

	"github.com/nats-io/go-nats"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges []string) ([]*types.Bridge, error) {

	nc, err := nats.Connect(connect)
	if err != nil {
		return nil, err
	}

	brs := []*types.Bridge{}

	for i, br := range bridges {
		pair := strings.SplitN(br, ":", 2)
		if len(pair) == 1 {
			pair = append(pair, pair[0])
		}

		ch := make(chan *nats.Msg, 64)
		sub, err := nc.ChanSubscribe(pair[0], ch)
		if err != nil {
			return nil, err
		}

		b := types.Bridge{ID: i, Channel: pair[1], Pipe: make(chan string)}
		side.Log.Printf("Bridge %d: side 'in' connect to channel (%s)", b.ID, pair[0])
		go reader(side, sub, ch, b)
		side.WG.Add(1)
		brs = append(brs, &b)
	}
	go disconnect(side, nc)
	side.WGControl.Add(1)
	return brs, nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges []*types.Bridge) error {

	nc, err := nats.Connect(connect)
	if err != nil {
		return err
	}
	for _, b := range bridges {
		side.Log.Printf("Bridge %d side 'out': connect to func (%s)", b.ID, b.Channel)
		go writer(side, nc, *b)
		side.WG.Add(1)
	}
	go disconnect(side, nc)
	side.WGControl.Add(1)
	return nil
}

func reader(side *types.Side, sub *nats.Subscription, ch chan *nats.Msg, br types.Bridge) {
	defer side.WG.Done()
	defer sub.Unsubscribe()
	for {
		select {
		case line := <-ch:
			br.Pipe <- string(line.Data[:])
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d side 'in' closed", br.ID)
			return
		}
	}
}

func writer(side *types.Side, nc *nats.Conn, br types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			nc.Publish(br.Channel, []byte(line))
			//	side.Abort <- br.ID
			//	return
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d side 'out' closed", br.ID)
			return
		}
	}
}

func disconnect(side *types.Side, nc *nats.Conn) {
	defer side.WGControl.Done()
	defer nc.Close()

	for {
		select {
		case <-side.Quit:
			side.WG.Wait()
			side.Log.Println("debug: DB disconnect")
			return
		}
	}
}
