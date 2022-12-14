package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/fioncat/wshare/pkg/log"
)

var (
	key []byte

	block cipher.Block
	gcm   cipher.AEAD
)

func Init(password string) error {
	sum := sha256.Sum256([]byte(password))
	key = sum[:32]

	var err error
	block, err = aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to validate aes key: %v", err)
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to init gcm: %v", err)
	}

	return nil
}

func Encrypt(data []byte) []byte {
	if len(key) == 0 {
		return data
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())

	// populates our nonce with a cryptographically secure
	// random sequence
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		// This should not be triggered in normal case
		log.Get().Warnf("internal: failed to generate random sequence: %v", err)
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

func Decrypt(data []byte) ([]byte, error) {
	if len(key) == 0 {
		return data, nil
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("no an aes data")
	}

	var nonce []byte
	nonce, data = data[:nonceSize], data[nonceSize:]
	src, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}
	return src, nil
}
