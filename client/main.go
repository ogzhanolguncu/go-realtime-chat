package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
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

	handleNameSet(conn, reader)
	readMessagesFromServer(conn)
	sendMessagesToServer(reader, conn)
}

func printHeader() {
	fmt.Printf("\n\n")
	fmt.Println("---------CHATROOM--------")
	fmt.Println("-------------------------")
}

func readMessagesFromServer(conn net.Conn) {
	go func() {
		for {
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			handleIncomingMessage(message, askForInput)
		}
	}()
}

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

func handleNameSet(conn net.Conn, reader *bufio.Reader) {
	serverReader := bufio.NewReader(conn)

	message, err := serverReader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading from server:", err)
	}
	if strings.TrimSpace(message) != "USERNAME_REQUIRED" {
		log.Fatal("Expected USERNAME_REQUIRED message from server")
	}

	retries := 0
	for retries < 3 {
		fmt.Print("Enter your username: ")
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading username input:", err)
		}
		nameInput = strings.TrimSpace(nameInput)
		conn.Write([]byte(nameInput + "\n"))

		message, err = serverReader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading from server:", err)
		}
		decodedMessage, _ := decodeMessage(message)

		if strings.Contains(decodedMessage.sysStatus, "fail") {
			colorifyAndFormatContent(decodedMessage)
			retries++
			continue
		}

		if strings.Contains(decodedMessage.sysStatus, "success") {
			colorifyAndFormatContent(decodedMessage)
			name = strings.Split(strings.TrimSpace(decodedMessage.content), ":")[1]
			return
		}

		fmt.Println("Unexpected response from server. Please try again.")
		retries++
	}

	fmt.Println("Max retries reached. Exiting.")
	conn.Close() // Close the connection
}

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

func sendWhisperMessage(conn net.Conn, recipient, msg string) {
	_, err := conn.Write([]byte(fmt.Sprintf("WHISPER#%s#%s#%s\n", name, msg, recipient)))
	if err != nil {
		log.Fatal("Error sending whisper message:", err)
	}
}

func sendGroupMessage(conn net.Conn, text string) {
	_, err := conn.Write([]byte(fmt.Sprintf("GROUP_MESSAGE#%s#%s\n", name, text)))
	if err != nil {
		log.Fatal("Error sending group message:", err)
	}
}

func askForInput() {
	coloredPrompt := color.ColorifyWithTimestamp("You:", color.Yellow)
	fmt.Printf("%s ", coloredPrompt)
}
