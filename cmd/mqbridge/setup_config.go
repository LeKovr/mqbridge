package main

import (
	"errors"

	"github.com/jessevdk/go-flags"
)

// -----------------------------------------------------------------------------

// Flags defines local application flags
type Flags struct {
	Version bool `long:"version"                       description:"Show version and exit"`
	Debug   bool `long:"debug"                         description:"Show debug data"`
}

var (
	// ErrGotHelp returned after showing requested help
	ErrGotHelp = errors.New("help printed")
	// ErrBadArgs returned after showing command args error message
	ErrBadArgs = errors.New("option error printed")
)

// SetupConfig loads flags from args (if given) or command flags and ENV otherwise
func SetupConfig(args ...string) (*Config, error) {
	cfg := &Config{}
	p := flags.NewParser(cfg, flags.Default) //  HelpFlag | PrintErrors | PassDoubleDash

	var err error
	if len(args) == 0 {
		_, err = p.Parse()
	} else {
		_, err = p.ParseArgs(args)
	}
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			return nil, ErrGotHelp
		}
		return nil, ErrBadArgs
	}
	return cfg, nil
}
