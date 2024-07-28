package internal

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

// Client sends base64 encoded publicKey
// Server converts base64 to publicKey
// Server encrypts groupChatKey using publicKey
// Server sends encrypted groupChatKey
// Client reads encrypted groupChatKey
// Client decrypts encryptedData and gets groupChatKey

// Each user (e.g., Alice, Bob, Charlie) connects to the server and provides their public key.
// The server encrypts the group key with each user’s public key and sends it to them.
// Users decrypt the group key using their private keys.
func (c *Client) FetchGroupChatKey() error {
	serverReader := bufio.NewReader(c.conn)
	message, err := c.prepareEncryptionPayload(c.publicKey)
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(message))
	if err != nil {
		return err
	}

	serverResp, err := serverReader.ReadString('\n')
	if err != nil {
		return err
	}
	decodedMsg, err := c.decodeFn(serverResp)
	if err != nil {
		return err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(decodedMsg.EncryptedKey)
	if err != nil {
		return fmt.Errorf("failed to decode base64 group chat key: %w", err)
	}

	groupChatKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, keyBytes, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt group chat key: %w", err)
	}

	c.groupChatKey = string(groupChatKey)
	return nil
}

func (c *Client) prepareEncryptionPayload(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)

	return c.encodeFn(protocol.Payload{
		MessageType: protocol.MessageTypeENC, EncryptedKey: publicKeyStr,
	}), nil
}
