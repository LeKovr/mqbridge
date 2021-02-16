package main

import "github.com/go-logr/logr"

// Shutdown runs exit after deferred cleanups have run
func Shutdown(exitFunc func(code int), e error, log logr.Logger) {
	if e != nil {
		var code int
		switch e {
		case ErrGotHelp:
			code = 3
		case ErrBadArgs:
			code = 2
		default:
			log.Error(e, "Run error")
			code = 1
		}
		exitFunc(code)
	}
}
