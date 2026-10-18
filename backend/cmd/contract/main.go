package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/SemiuAdesina/nuvro/backend/internal/contract"
)

func main() {
	source := flag.String("source", "../contracts/openapi/v1/openapi.yaml", "OpenAPI source")
	output := flag.String("output", "", "Optional bundle output")
	flag.Parse()
	bundle, err := contract.Bundle(context.Background(), *source)
	if err != nil {
		log.Fatal(err)
	}
	if *output != "" {
		if err := os.WriteFile(*output, bundle, 0600); err != nil {
			log.Fatal(err)
		}
	}
}
