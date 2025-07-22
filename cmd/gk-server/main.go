package main

import (
	"github.com/nekr0z/gk/internal/cli"
	server "github.com/nekr0z/gk/internal/server/cli"
)

func main() {
	cli.Execute(server.RootCmd())
}
