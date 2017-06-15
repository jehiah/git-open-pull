package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	settings, err := LoadSettings(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%#v", settings)
}
