package main

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("failed to connect server: %v", err)
	}

	if err := client.SetUsername(); err != nil {
		log.Fatalf("failed to set username: %v", err)
	}

	incomingChan := make(chan protocol.Payload)
	outgoingChan := make(chan string)
	errChan := make(chan error, 1) // Buffered channel to prevent goroutine leak
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2) // Add 2 for readMessages and sendMessages goroutines

	go func() {
		defer wg.Done()
		client.ReadMessages(incomingChan, errChan, done)
	}()

	go func() {
		defer wg.Done()
		client.SendMessages(outgoingChan, done)
	}()

	// Signal to stop goroutines
	close(done)

	// Wait for goroutines to finish
	wg.Wait()
}
