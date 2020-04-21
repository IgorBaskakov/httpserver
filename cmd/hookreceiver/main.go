package main

import (
	"log"
	"os"

	"github.com/IgorBaskakov/httpserver/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	internal.StartServer(port)
}
