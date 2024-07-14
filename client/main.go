package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	err := retry.Do(
		func() error {
			return runClient()
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			if err.Error() == "EOF" {
				err = fmt.Errorf("server is not responding")
			}
			fmt.Println(color.ColorifyWithTimestamp(fmt.Sprintf("Trying to reconnect, but : %v", err), color.Red))
		}),
	)

	if err != nil {
		log.Fatalf("Failed after max retries: %v", err)
	}
}

func runClient() error {
	config := Config{
		Port: 7007,
	}
	client, err := newClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.close()

	if err := client.connect(); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	client.printHeader()
	if err := client.setUsername(); err != nil {
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
		client.readMessages(incomingChan, errChan, done)
	}()

	go func() {
		defer wg.Done()
		client.sendMessages(outgoingChan, done)
	}()

	err = client.messageLoop(incomingChan, outgoingChan, errChan, done)

	// Signal to stop goroutines
	close(done)

	// Wait for goroutines to finish
	wg.Wait()

	return err
}

func (c *Client) messageLoop(incomingChan <-chan protocol.Payload, outgoingChan <-chan string, errChan <-chan error, done chan struct{}) error {
	for {
		select {
		case incMessage, ok := <-incomingChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			if incMessage.ContentType == protocol.MessageTypeWSP {
				c.lastWhispererFromGroupChat = incMessage.Sender
			}
			colorifyAndFormatContent(incMessage)
			askForInput()
		case outMessage, ok := <-outgoingChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			_, err := c.conn.Write([]byte(outMessage))
			if err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}
		case err, ok := <-errChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			return err
		case <-done:
			return nil // Done signal received, exit loop
		}
	}
}
