package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/ogzhanolguncu/go-chat/server/internal/server"
	"github.com/ogzhanolguncu/go-chat/server/utils"
)

const (
	port   = 7007
	dbName = "chat.db"
)

func main() {
	encoding := flag.Bool("encoding", false, "enable encoding")
	flag.Parse()

	dbPath := filepath.Join(utils.RootDir(), dbName)

	s, err := server.NewServer(port, dbPath, *encoding)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	}()

	log.Printf("Chat server starting on port %d\n", port)
	log.Printf("Encoding: %v\n", *encoding)

	s.Start()
}
