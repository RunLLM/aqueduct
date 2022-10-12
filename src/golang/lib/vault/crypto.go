package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
)

// encrypt uses `key` to encrypt `secrets`
func encrypt(secrets map[string]string, key string) ([]byte, error) {
	// generate an aes cipher using our 32 byte encryption key
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// Galois Counter Mode (GCM) is used symmetric key cryptographic block ciphers
	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// generate nonce with a random sequence
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// allocate a byte array for encrypted value
	dst := make([]byte, len(nonce))

	serialized, err := json.Marshal(secrets)
	if err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(dst, nonce, serialized, nil /* additionalData */)
	return encrypted, nil
}

// decrypt uses `key` to decrypt `data`
func decrypt(data []byte, key string) (map[string]string, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(plaintext, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
