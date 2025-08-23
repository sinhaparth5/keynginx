package main

import (
	"os"
	"github.com/sinhaparth5/keynginx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}