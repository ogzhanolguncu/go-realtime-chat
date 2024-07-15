package internal

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/ogzhanolguncu/go-chat/client/color"
	"github.com/ogzhanolguncu/go-chat/protocol"
)

func (c *Client) FetchActiveUsersAfterUsername() error {
	serverReader := bufio.NewReader(c.conn)
	message, err := handleActiveUsers("", "", "")
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(message))
	if err != nil {
		return err
	}

	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	serverResp, err := serverReader.ReadString('\n')
	c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}
	decodedMsg, err := protocol.DecodeMessage(serverResp)
	if err != nil {
		return err
	}
	// Active Users (25): Alice, Bob, Charlie [+22 more] (Use /users for full list)
	msg := formatActiveUsers(decodedMsg.ActiveUsers)
	fmt.Println(color.ColorifyWithTimestamp(msg, color.Blue))
	fmt.Println("")
	return nil
}

func formatActiveUsers(activeUsers []string) string {
	userCount := len(activeUsers)
	displayUsers := getFirstN(activeUsers, 3)

	var parts []string
	for _, user := range displayUsers {
		if user != "" {
			parts = append(parts, user)
		}
	}

	result := fmt.Sprintf("Active Users (%d): %s", userCount, strings.Join(parts, ", "))

	if userCount > 3 {
		result += fmt.Sprintf(" [+%d more]", userCount-3)
	}

	result += " (Use /users for full list)"

	return result
}

func getFirstN(arr []string, n int) []string {
	result := make([]string, n)
	for i := 0; i < n && i < len(arr); i++ {
		result[i] = arr[i]
	}
	return result
}
