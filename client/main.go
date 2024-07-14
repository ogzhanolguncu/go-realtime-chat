package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/joho/godotenv"
	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = retry.Do(
		func() error {
			return runClient()
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			if err.Error() == "EOF" {
				err = fmt.Errorf("server is not responding")
			}
			fmt.Println(color.ColorifyWithTimestamp(fmt.Sprintf("Trying to reconnect, but %v", err), color.Red))
		}),
	)

	if err != nil {
		log.Fatalf("Failed after max retries: %v", err)
	}
}

func runClient() error {
	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	client.PrintHeader()
	if err := client.SetUsername(); err != nil {
		return fmt.Errorf("failed to set username: %v", err)
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

	err = client.MessageLoop(incomingChan, outgoingChan, errChan, done)

	// Signal to stop goroutines
	close(done)

	// Wait for goroutines to finish
	wg.Wait()

	return err
}
