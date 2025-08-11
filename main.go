package main

import (
	"log"

	"r34-go/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
