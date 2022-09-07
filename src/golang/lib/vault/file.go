package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	FileVaultDir       = "vault/"
	filePermissionCode = 0o664
)

type fileVault struct {
	fileConfig *FileConfig
}

func NewFileVault(conf *FileConfig) (*fileVault, error) {
	return &fileVault{fileConfig: conf}, nil
}

func (f *fileVault) Config() Config {
	return f.fileConfig
}

func (f *fileVault) Put(ctx context.Context, name string, secrets map[string]string) error {
	serialized, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher([]byte(f.fileConfig.EncryptionKey))
	if err != nil {
		return err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// the WriteFile method returns an error if unsuccessful
	return os.WriteFile(f.getFullPath(name), gcm.Seal(nonce, nonce, serialized, nil), filePermissionCode)
}

func (f *fileVault) Get(ctx context.Context, name string) (map[string]string, error) {
	ciphertext, err := os.ReadFile(f.getFullPath(name))
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher([]byte(f.fileConfig.EncryptionKey))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
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

func (f *fileVault) Delete(ctx context.Context, name string) error {
	return os.Remove(f.getFullPath(name))
}

func (f *fileVault) getFullPath(key string) string {
	return fmt.Sprintf("%s/%s", f.fileConfig.Directory, key)
}
