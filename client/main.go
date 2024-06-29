package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const port = 7007

var (
	name string
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n\n")
	fmt.Println("---------CHATROOM--------")
	fmt.Println("-------------------------")

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, name = askForUsername("-> NAME: ")

	//Clears screen
	fmt.Print("\033[H\033[2J")

	// Starts a goroutine to read messages from the server
	go func() {
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			printIncomingMessage(message)
		}
	}()

	// Main loop to send messages to the server
	for {
		askForInput()
		text, _ := reader.ReadString('\n')
		// Convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("quit", text) == 0 {
			break
		}

		_, err = conn.Write([]byte(fmt.Sprintf("%s: %s\n", name, text)))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func askForUsername(s string) (bool, string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.TrimSpace(response)

		if response != "" {
			return true, response
		}

		fmt.Println("Input cannot be empty. Please try again.")
	}
}

func printIncomingMessage(message string) {
	fmt.Printf("\r-> %s", message)
	askForInput()
}

func askForInput() {
	fmt.Printf("-> %s: ", name)
}
