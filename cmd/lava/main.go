// Copyright 2023 Adevinta

// Lava runs Vulcan checks locally.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jroimartin/clilog"

	"github.com/adevinta/lava/cmd/lava/internal/base"
	"github.com/adevinta/lava/cmd/lava/internal/help"
	"github.com/adevinta/lava/cmd/lava/internal/run"
)

func init() {
	base.Commands = []*base.Command{
		run.CmdRun,
	}
}

func main() {
	h := clilog.NewCLIHandler(os.Stderr, &clilog.HandlerOptions{Level: base.LogLevel})
	slog.SetDefault(slog.New(h))

	flag.Usage = help.PrintUsage
	flag.Parse() //nolint:errcheck

	args := flag.Args()
	if len(args) < 1 {
		help.PrintUsage()
		os.Exit(2)
	}

	if args[0] == "help" {
		help.Help(args[1:])
		return
	}

	for _, cmd := range base.Commands {
		cmd.Flag.Usage = cmd.Usage
		if cmd.Name() == args[0] {
			cmd.Flag.Parse(args[1:]) //nolint:errcheck
			args = cmd.Flag.Args()
			if err := cmd.Run(args); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "lava: unknown command %q\nRun 'lava help' for usage.\n", args[0])
	os.Exit(2)
}