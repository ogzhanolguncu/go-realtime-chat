package main

import (
	"log"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

func main() {
	config := Config{
		Port: 7007,
	}

	client, err := newClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.close()

	if err := client.connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client.printHeader()

	if err := client.setUsername(); err != nil {
		log.Fatalf("Failed to set username: %v", err)
	}

	incomingChan := make(chan protocol.Payload)
	outgoingChan := make(chan string)
	go client.readMessages(incomingChan)
	go client.sendMessages(outgoingChan)

	for {
		select {
		case incMessage := <-incomingChan:
			// This is required for /reply function to work.
			if incMessage.ContentType == protocol.MessageTypeWSP {
				client.lastWhispererFromGroupChat = incMessage.Sender
			}
			colorifyAndFormatContent(incMessage)
			// When message received from server we append You: right after it.
			askForInput()
		case outMessage := <-outgoingChan:
			_, err := client.conn.Write([]byte(outMessage))
			if err != nil {
				log.Fatal("Error sending group message:", err)
			}
		}
	}
}
