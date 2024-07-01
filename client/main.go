package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
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
			printIncomingMessage(message)
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
	if strings.HasPrefix(strings.TrimSpace(message), "USERNAME_SET_SUCCESSFULLY#") {
		name = strings.Split(strings.TrimSpace(message), "#")[1]
		fmt.Printf("\033[32mUsername successfully set to: %s\033[0m\n\n", name)
	} else {
		log.Fatal("Expected USERNAME_SET_SUCCESSFULLY message from server")
	}
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

func printIncomingMessage(message string) {
	timestamp := time.Now().Format("[15:04]")

	if strings.HasPrefix(message, "System:") {
		fmt.Printf("\r\033[36m%s %s\033[0m\n", timestamp, message) // Cyan for system messages
	} else if strings.HasPrefix(message, "Whisper from") {
		fmt.Printf("\r\033[35m%s %s\033[0m\n", timestamp, message) // Purple for whisper messages
	} else {
		fmt.Printf("\r\033[34m%s %s\033[0m\n", timestamp, message) // Blue for group messages
	}

	askForInput()
}

func askForInput() {
	fmt.Printf("\033[33m[%.2d:%.2d] You:\033[0m ", time.Now().Hour(), time.Now().Minute())
}
