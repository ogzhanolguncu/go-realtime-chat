package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) SendMessages(outgoingChan chan<- string, done <-chan struct{}) {
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
			return
		}

		var message string
		if strings.HasPrefix(text, "/reply") {
			message, err = c.sendReply(text)
		} else if strings.HasPrefix(text, "/whisper") {
			message, err = c.sendWhisper(text)
		} else {
			message = c.sendPublicMessage(text)
		}

		if err != nil {
			log.Println("Error preparing message:", err)
			continue
		}

		select {
		case outgoingChan <- message:
			// Message sent successfully
		case <-done:
			return
		}
	}
}

func (c *Client) sendReply(text string) (message string, err error) {
	return c.sendWhisper(fmt.Sprintf("/whisper %s %s", c.lastWhispererFromGroupChat, strings.TrimSpace(strings.Split(text, "/reply")[1])))
}

func (c *Client) sendWhisper(text string) (message string, err error) {
	re := regexp.MustCompile(`^\/whisper\s+(\S+)\s+(.*)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		recipient := matches[1]
		msg := matches[2]
		return protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeWSP, Recipient: recipient, Sender: c.name, Content: msg}), nil
	} else {
		fmt.Println("Invalid whisper command format")
		return "", nil
	}
}

func (c *Client) sendPublicMessage(rawInput string) (message string) {
	return protocol.EncodeMessage(protocol.Payload{ContentType: protocol.MessageTypeMSG, Sender: c.name, Content: rawInput})
}

//RECEIVER

func (c *Client) ReadMessages(incomingChan chan<- protocol.Payload, errChan chan<- error, done <-chan struct{}) {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			select {
			case errChan <- err:
			case <-done:
			}
			return
		}
		payload, err := protocol.DecodeMessage(message)
		if err != nil {
			fmt.Print(color.ColorifyWithTimestamp(err.Error(), color.Red))
			continue
		}
		select {
		case incomingChan <- payload:
		case <-done:
			return
		}
	}
}

//MESSAGE LOOP

func (c *Client) MessageLoop(incomingChan <-chan protocol.Payload, outgoingChan <-chan string, errChan <-chan error, done chan struct{}) error {
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
