// Command papercuts records workflow friction encountered while doing other work.
package main

import (
	"context"
	"os"

	"golang.org/x/term"

	"github.com/Whamp/papercuts/internal/buildinfo"
	"github.com/Whamp/papercuts/internal/cli"
	"github.com/Whamp/papercuts/internal/papercuts"
)

func main() {
	service := papercuts.NewService()
	exitCode := cli.Run(
		context.Background(),
		os.Args[1:],
		cli.IO{
			Stdin:      os.Stdin,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			StdinIsTTY: term.IsTerminal(int(os.Stdin.Fd())),
		},
		service,
		buildinfo.Current(),
	)
	os.Exit(exitCode)
}
