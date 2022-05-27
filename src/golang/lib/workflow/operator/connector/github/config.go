package github

import "encoding/gob"

type ManagerType string

const NoopManagerType ManagerType = "noop"

type ManagerConfig interface {
	Type() ManagerType
}

type UnimplementedManagerConfig struct{}

func (*UnimplementedManagerConfig) Type() ManagerType {
	return NoopManagerType
}

func RegisterGobTypes() {
	gob.Register(&UnimplementedManagerConfig{})
}

func init() {
	RegisterGobTypes()
}
