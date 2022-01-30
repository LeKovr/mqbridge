package types_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/LeKovr/mqbridge/types"
)

func Example_printer() {
	ctx, cancel := context.WithCancel(context.Background())
	epa := types.NewBlankEndPointAttr(ctx)
	pipe := make(chan string)

	go epa.Printer(logr.Discard(), pipe)
	pipe <- "test row"
	time.Sleep(100 * time.Millisecond)
	cancel()
	epa.WG.Wait()
	// Output:
	// test row
}
