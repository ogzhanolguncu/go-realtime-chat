package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/ogzhanolguncu/go-chat/protocol"
)

// Client sends base64 encoded publicKey
// Server converts base64 to publicKey
// Server encrypts groupChatKey using publicKey
// Server sends encrypted groupChatKey
// Client reads encrypted
// Client decrypts encryptedData and gets groupChatKey

// Each user (e.g., Alice, Bob, Charlie) connects to the server and provides their public key.
// The server encrypts the group key with each user’s public key and sends it to them.
// Users decrypt the group key using their private keys.
func (c *Client) FetchGroupChatKey(errorChan chan error) {
	message, err := c.prepareEncryptionPayload(c.publicKey)
	if err != nil {
		errorChan <- err
	}
	_, err = c.conn.Write([]byte(message))
	if err != nil {
		errorChan <- err
	}
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
