package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

const port = 7007

var (
	name string
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	printHeader()

	conn, internalError := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if internalError != nil {
		log.Fatal(internalError)
	}
	defer conn.Close()

	// Waits for "USERNAME_REQUIRED" from the server
	// Asks for and send username to the server
	// Waits for "USERNAME_SET_SUCCESSFULLY" message from the server
	handleNameSet(conn, reader)

	// Start a goroutine to read messages from the server
	readMessagesFromServer(conn)

	// Main loop to send messages to the server
	sendMessagesToServer(reader, conn)
}

func printHeader() {
	fmt.Printf("\n\n")
	fmt.Println("---------CHATROOM--------")
	fmt.Println("-------------------------")
}

func readMessagesFromServer(conn net.Conn) {
	// Start a goroutine to read messages from the server
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
}

// sendMessagesToServer reads user input and sends messages to the server
func sendMessagesToServer(reader *bufio.Reader, conn net.Conn) {
	for {
		askForInput()
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "quit" {
			break
		}

		if strings.HasPrefix(text, "/w ") {
			handleWhisperCommand(conn, text)
		} else {
			sendGroupMessage(conn, text)
		}
	}
}

// handleNameSet manages the username setup process with the server
func handleNameSet(conn net.Conn, reader *bufio.Reader) {
	serverReader := bufio.NewReader(conn)

	// Wait for "USERNAME_REQUIRED" from the server
	message, err := serverReader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading from server:", err)
	}
	if strings.TrimSpace(message) != "USERNAME_REQUIRED" {
		log.Fatal("Expected USERNAME_REQUIRED message from server")
	}

	// Ask for and send username to the server
	fmt.Print("Enter your username: ")
	nameInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading username input:", err)
	}
	nameInput = strings.TrimSpace(nameInput)
	conn.Write([]byte(nameInput + "\n"))

	// Wait for "USERNAME_SET_SUCCESSFULLY" message from the server
	message, err = serverReader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading from server:", err)
	}
	if strings.HasPrefix(strings.TrimSpace(message), "USERNAME_SET_SUCCESSFULLY#") {
		name = strings.Split(strings.TrimSpace(message), "#")[1]
		fmt.Printf("Username successfully set to: %s\n\n", name)
	} else {
		log.Fatal("Expected USERNAME_SET_SUCCESSFULLY message from server")
	}
}

// handleWhisperCommand processes and sends a whisper message
func handleWhisperCommand(conn net.Conn, text string) {
	re := regexp.MustCompile(`^\/w\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		sendWhisperMessage(conn, recipient, msg)
	} else {
		fmt.Println("Invalid whisper command format")
	}
}

// sendWhisperMessage sends a whisper message to the server
func sendWhisperMessage(conn net.Conn, recipient, msg string) {
	_, err := conn.Write([]byte(fmt.Sprintf("WHISPER#%s#%s#%s\n", name, msg, recipient)))
	if err != nil {
		log.Fatal("Error sending whisper message:", err)
	}
}

// sendGroupMessage sends a group message to the server
func sendGroupMessage(conn net.Conn, text string) {
	_, err := conn.Write([]byte(fmt.Sprintf("GROUP_MESSAGE#%s#%s\n", name, text)))
	if err != nil {
		log.Fatal("Error sending group message:", err)
	}
}

func printIncomingMessage(message string) {
	fmt.Printf("\r-> %s", message)
	askForInput()
}

func askForInput() {
	fmt.Printf("\033[0;37m-> ME: ")
}
