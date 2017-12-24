package pg

import (
	"github.com/go-pg/pg"

	"github.com/LeKovr/mqbridge/types"
)

// Listen starts all listening goroutines
func Listen(side *types.Side, connect string, bridges types.Bridges) error {

	side.Log.Printf("Connect PG producer: %s", connect)
	opts, err := pg.ParseURL(connect)
	if err != nil {
		return err
	}
	db := pg.Connect(opts)

	for _, br := range bridges {
		ln := db.Listen(br.In)
		side.Log.Printf("Bridge %d: producer connect to channel (%s)", br.ID, br.In)
		go reader(side, ln, br)
		side.WG.Add(1)
	}
	go disconnect(side, db)
	side.WGControl.Add(1)
	return nil
}

// Notify starts all notify goroutines
func Notify(side *types.Side, connect string, bridges types.Bridges) error {

	side.Log.Printf("Connect PG consumer: %s", connect)
	opts, err := pg.ParseURL(connect)
	if err != nil {
		return err
	}
	db := pg.Connect(opts)
	for _, br := range bridges {
		side.Log.Printf("Bridge %d consumer: connect to func (%s)", br.ID, br.Out)
		go writer(side, db, br)
		side.WG.Add(1)
	}
	go disconnect(side, db)
	side.WGControl.Add(1)
	return nil
}

func reader(side *types.Side, ln *pg.Listener, br *types.Bridge) {
	defer side.WG.Done()
	defer ln.Close()
	ch := ln.Channel()
	for {
		side.Log.Println("debug: BRIN ", br.ID, " START")
		select {
		case line := <-ch:
			side.Log.Println("debug: BRIN ", br.ID, " ", line.Payload)
			br.Pipe <- line.Payload
		case <-side.Quit:
			side.Log.Printf("debug: Bridge %d producer closed", br.ID)
			return
		}
	}
}

func writer(side *types.Side, db *pg.DB, br *types.Bridge) {
	defer side.WG.Done()
	for {
		select {
		case line := <-br.Pipe:
			_, err := db.Exec("SELECT "+br.Out+"(?)", line)
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
