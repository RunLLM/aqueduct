package vault

import "encoding/gob"

type Type string

const (
	FileType Type = "file"
)

type Config interface {
	Type() Type
}

type FileConfig struct {
	Directory string `json:"directory" yaml:"directory"`
	// Length can be 128, 192, or 256 bits
	EncryptionKey string `json:"encryption_key" yaml:"encryptionKey"`
}

func (*FileConfig) Type() Type {
	return FileType
}

func RegisterGobTypes() {
	gob.Register(&FileConfig{})
}

func init() {
	RegisterGobTypes()
}
