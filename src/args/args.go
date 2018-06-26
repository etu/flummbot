package args

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

type Args struct {
	ConfigFile string
	Debug      bool
}

func Parse() Args {
	parser := argparse.NewParser(
		os.Args[0],
		"Run flummbot",
	)

	// Add flags
	configFile := parser.String("c", "config", &argparse.Options{
		Required: true,
		Help:     "Path to config file to use",
	})

	debug := parser.Flag("d", "debug", &argparse.Options{
		Help: "Print debug info if this flag is present",
	})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	return Args{*configFile, *debug}
}
