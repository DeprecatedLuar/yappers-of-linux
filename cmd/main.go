package main

import (
	"os"

	"yappers-of-linux/internal/commands"
)

func main() {
	commands.Parse(os.Args)
}
