package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/inky-fetch/fetchapp"
)

func main() {
	exitcode.Exit(fetchapp.CLI(os.Args[1:]))
}
