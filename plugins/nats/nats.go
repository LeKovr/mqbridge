package nats

import (
	"github.com/nats-io/go-nats"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges types.Bridges) error {

	side.Log.Printf("Connect NATS producer: %s", connect)
	nc, err := nats.Connect(connect)
	if err != nil {
		return err
	}

	for _, br := range bridges {
		ch := make(chan *nats.Msg, 64)
		sub, err := nc.ChanSubscribe(br.In, ch)
		if err != nil {
			return err
		}
		side.Log.Printf("Bridge %d: producer connect to channel (%s)", br.ID, br.In)
		go reader(side, sub, ch, br)
		side.WG.Add(1)
	}
	go disconnect(side, nc)
	side.WGControl.Add(1)
	return nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges types.Bridges) error {

	side.Log.Printf("Connect NATS consumer: %s", connect)
	nc, err := nats.Connect(connect)
	if err != nil {
		return err
	}
	for _, br := range bridges {
		side.Log.Printf("Bridge %d consumer: connect to channel (%s)", br.ID, br.Out)
		go writer(side, nc, br)
		side.WG.Add(1)
	}
	go disconnect(side, nc)
	side.WGControl.Add(1)
	return nil
}

func reader(side *types.Side, sub *nats.Subscription, ch chan *nats.Msg, br *types.Bridge) {
	defer side.WG.Done()
	defer sub.Unsubscribe()
	for {
		select {
		case line := <-ch:
			str := string(line.Data[:])
			side.Log.Println("debug: BRIN ", br.ID, " ", str)
			br.Pipe <- str
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d producer closed", br.ID)
			return
		}
	}
}

func writer(side *types.Side, nc *nats.Conn, br *types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			err := nc.Publish(br.Out, []byte(line))
			if err != nil {
				side.Log.Printf("warn: Bridge %d consumer error: %v", br.ID, err.Error())
				//	side.Abort <- br.ID
				//	return
			}
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d consumer closed", br.ID)
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
			side.Log.Println("debug: NATS disconnect")
			return
		}
	}
}
