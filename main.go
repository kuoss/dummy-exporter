package main

import (
	"log"
	"os"

	"github.com/kuoss/dummy-exporter/pkg/app"
)

const Version = "development"

func main() {
	log.SetOutput(os.Stdout)
	if err := app.Start(Version); err != nil {
		log.Fatalf("dummy-exporter exited with error: %v", err)
	}
}
