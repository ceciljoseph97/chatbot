package ipc

import (
	"io"

	"github.com/natefinch/npipe"
)

type IPCWindows struct {
	PipeName string
}

func NewIPCWindows(pipeName string) *IPCWindows {
	return &IPCWindows{PipeName: pipeName}
}
func (ipc *IPCWindows) Listen() (io.ReadWriteCloser, error) {
	listener, err := npipe.Listen(ipc.PipeName)
	if err != nil {
		return nil, err
	}
	conn, err := listener.Accept()
	if err != nil {
		listener.Close()
		return nil, err
	}
	listener.Close()
	return conn, nil
}

func (ipc *IPCWindows) Connect() (io.ReadWriteCloser, error) {
	return npipe.Dial(ipc.PipeName)
}
