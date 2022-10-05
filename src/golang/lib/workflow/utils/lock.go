package utils

import "github.com/dropbox/godropbox/sys/filelock"

// executionLock is the name of the shared filelock for blocking workflow run execution
const executionLock = "ExecutionLock"

// NewExecutionLock returns a new workflow execution mutex
func NewExecutionLock() *filelock.FileLock {
	return filelock.New(executionLock)
}
