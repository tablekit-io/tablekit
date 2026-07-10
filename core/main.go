package main

import (
	"core/cli"
	"core/logging"
)

func main() {
	// Install zerolog and its redirects before anything else runs, so every
	// subcommand — and the goose migrations that fire during services.New — logs
	// as structured JSON from the first line.
	logging.Init()
	cli.Execute()
}
