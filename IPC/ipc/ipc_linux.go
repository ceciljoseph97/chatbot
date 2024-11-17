package ipc

import (
	"fmt"
	"io"
	"net"
	"os"
)

type IPCLinux struct {
	SocketPath string
}

func NewIPCLinux(socketPath string) *IPCLinux {
	return &IPCLinux{SocketPath: socketPath}
}

func (ipc *IPCLinux) Listen() (io.ReadWriteCloser, error) {
	if _, err := os.Stat(ipc.SocketPath); err == nil {
		os.Remove(ipc.SocketPath)
	}

	listener, err := net.Listen("unix", ipc.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on unix socket %s: %v", ipc.SocketPath, err)
	}

	conn, err := listener.Accept()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to accept connection on unix socket: %v", err)
	}

	listener.Close()

	return conn, nil
}

func (ipc *IPCLinux) Connect() (io.ReadWriteCloser, error) {
	return net.Dial("unix", ipc.SocketPath)
}
