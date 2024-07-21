package protocol

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func Encrypt(msg []byte, key string) (string, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], msg)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encryptedMsg string, key string) ([]byte, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	decodedCiphertext, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return nil, err
	}

	if len(decodedCiphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := decodedCiphertext[:aes.BlockSize]
	decodedCiphertext = decodedCiphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(decodedCiphertext, decodedCiphertext)

	return decodedCiphertext, nil
}
