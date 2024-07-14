package main

import (
	"fmt"
	"log"
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
			fmt.Println(color.ColorifyWithTimestamp(fmt.Sprintf("Retry attempt %d: %v", n+1, err), color.Red))
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
	errChan := make(chan error)

	go client.readMessages(incomingChan, errChan)
	go client.sendMessages(outgoingChan)

	return client.messageLoop(incomingChan, outgoingChan, errChan)
}

func (c *Client) messageLoop(incomingChan chan protocol.Payload, outgoingChan chan string, errChan chan error) error {
	for {
		select {
		case incMessage := <-incomingChan:
			if incMessage.ContentType == protocol.MessageTypeWSP {
				c.lastWhispererFromGroupChat = incMessage.Sender
			}
			colorifyAndFormatContent(incMessage)
			askForInput()
		case outMessage := <-outgoingChan:
			_, err := c.conn.Write([]byte(outMessage))
			if err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}
		case err := <-errChan:
			return err
		}
	}
}
