package main

import (
	"os"
)

func main() {
	command := NewZuildCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
