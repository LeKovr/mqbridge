package types_test

import (
	"time"

	"github.com/LeKovr/mqbridge/types"
)

func Example_printer() {
	epa := types.NewBlankEndPointAttr()
	pipe := make(chan string)
	go epa.Printer(epa.Log, pipe)
	pipe <- "test row"
	time.Sleep(100 * time.Millisecond)
	close(epa.Quit)
	epa.WG.Wait()
	// Output:
	// test row
}
