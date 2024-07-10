package main

import (
	"log"
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

	go client.readMessages()
	client.sendMessages()
}
