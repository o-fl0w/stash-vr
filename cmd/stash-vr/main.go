//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi.yaml https://scripts01.handyfeeling.com/api/script/index/v0/spec

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
