package main

import (
	"github.com/nekr0z/gk/internal/cli"
	gk "github.com/nekr0z/gk/internal/manager/cli"
)

func main() {
	cli.Execute(gk.RootCmd())
}
