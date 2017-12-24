package pg

import (
	"strings"

	"github.com/go-pg/pg"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges []string) ([]*types.Bridge, error) {

	opts, err := pg.ParseURL(connect)
	if err != nil {
		return nil, err
	}
	db := pg.Connect(opts)

	brs := []*types.Bridge{}

	for i, br := range bridges {
		pair := strings.SplitN(br, ":", 2)
		if len(pair) == 1 {
			pair = append(pair, pair[0])
		}
		ln := db.Listen(pair[0])
		b := types.Bridge{ID: i, Channel: pair[1], Pipe: make(chan string)}
		side.Log.Printf("Bridge %d: producer connect to channel (%s)", b.ID, pair[0])
		go reader(side, ln, b)
		side.WG.Add(1)
		brs = append(brs, &b)
	}
	go disconnect(side, db)
	side.WGControl.Add(1)
	return brs, nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges []*types.Bridge) error {

	opts, err := pg.ParseURL(connect)
	if err != nil {
		return err
	}
	db := pg.Connect(opts)
	for _, b := range bridges {
		side.Log.Printf("Bridge %d consumer: connect to func (%s)", b.ID, b.Channel)
		go writer(side, db, *b)
		side.WG.Add(1)
	}
	go disconnect(side, db)
	side.WGControl.Add(1)
	return nil
}

func reader(side *types.Side, ln *pg.Listener, br types.Bridge) {
	defer side.WG.Done()
	defer ln.Close()
	ch := ln.Channel()
	for {
		select {
		case line := <-ch:
			br.Pipe <- line.Payload
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d producer closed", br.ID)
			return
		}
	}
}

func writer(side *types.Side, db *pg.DB, br types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			_, err := db.Exec("SELECT "+br.Channel+"(?)", line)
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

func disconnect(side *types.Side, db *pg.DB) {
	defer side.WGControl.Done()
	defer db.Close()

	for {
		select {
		case <-side.Quit:
			side.WG.Wait()
			side.Log.Println("debug: DB disconnect")
			return
		}
	}
}
