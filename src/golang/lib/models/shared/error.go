package shared

import "fmt"

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

func (e *Error) Message() string {
	errCtxMsg := ""
	if e.Context != "" {
		errCtxMsg = fmt.Sprintf("\nContext:\n%s", e.Context)
	}

	return fmt.Sprintf("%s%s", e.Tip, errCtxMsg)
}
