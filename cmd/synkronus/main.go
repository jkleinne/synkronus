// File: cmd/synkronus/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	app, err := newApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	Execute(app)
}
