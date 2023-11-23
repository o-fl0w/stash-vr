//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

package main

import (
	"log"
	"stash-vr/cmd/stash-vr/internal"
	"stash-vr/internal/interrupt"
)

func main() {
	if err := internal.Run(interrupt.Context()); err != nil {
		log.Fatal("Application EXIT with ERROR", err)
	}
}
