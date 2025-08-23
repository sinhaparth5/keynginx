package main

import (
	"github.com/sinhaparth5/keynginx/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
