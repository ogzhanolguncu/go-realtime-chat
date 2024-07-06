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

func (c *Client) sendMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

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
			c.sendWhisper(conn, text)
		} else {
			c.sendPublicMessage(conn, text)
		}
	}
}

func (c *Client) sendWhisper(conn net.Conn, text string) {
	re := regexp.MustCompile(`^\/w\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		_, err := conn.Write([]byte(fmt.Sprintf("WHISPER#%s#%s#%s\n", c.name, msg, recipient)))
		if err != nil {
			log.Fatal("Error sending whisper message:", err)
		}
	} else {
		fmt.Println("Invalid whisper command format")
	}
}

func (c *Client) sendPublicMessage(conn net.Conn, text string) {
	_, err := conn.Write([]byte(fmt.Sprintf("GROUP_MESSAGE#%s#%s\n", c.name, text)))
	if err != nil {
		log.Fatal("Error sending group message:", err)
	}
}
