package main

import (
	"pnas/cmd"
	_ "pnas/migrations" // Import migrations to register them
)

func main() {
	cmd.Execute()
}
