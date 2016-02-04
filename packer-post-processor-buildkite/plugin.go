package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	server.RegisterPostProcessor(new(PostProcessor))
	server.Serve()
}
