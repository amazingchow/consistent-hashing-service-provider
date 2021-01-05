package error

import (
	"errors"
	"fmt"
)

var (
	ErrBadAddress        = errors.New("bad address")
	ErrNoAvailableLeader = errors.New("no available leader")
	ErrNotReadyWorker    = errors.New("not ready worker")
	ErrNoAvailableNode   = errors.New("no available node")
)

type NoNodeError struct {
	uuid string
}

func (e *NoNodeError) Error() string {
	return fmt.Sprintf("no such node (uuid:%s)", e.uuid)
}

func NewNoNodeError(uuid string) *NoNodeError {
	return &NoNodeError{uuid: uuid}
}

type NodeExistsError struct {
	uuid string
}

func (e *NodeExistsError) Error() string {
	return fmt.Sprintf("node (uuid:%s) exists", e.uuid)
}

func NewNodeExistsError(uuid string) *NodeExistsError {
	return &NodeExistsError{uuid: uuid}
}
