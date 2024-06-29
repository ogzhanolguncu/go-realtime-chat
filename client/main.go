package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const PORT = 7007

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n\n")
	fmt.Println("-----CHAT CLIENT-----")
	fmt.Println("---------------------")

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(err)
	}
	_, name := askForUsername("-> NAME: ")
	defer conn.Close()

	for {
		fmt.Printf("-> %s: ", name)
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("quit", text) == 0 {
			break
		}

		// Send the message to the server
		_, err = conn.Write([]byte(fmt.Sprintf("%s: %s\n", name, text)))
		if err != nil {
			log.Fatal(err)
		}

	}

	for {
		// Read the response from the server
		_, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Print("-> " + message)
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

		response = strings.ToLower(strings.TrimSpace(response))

		if response != "" {
			return true, response
		}

		fmt.Println("Input cannot be empty. Please try again.")
	}
}
