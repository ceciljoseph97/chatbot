package ipc

import "io"

type IPC interface {
	Listen() (io.ReadWriteCloser, error)
	Connect() (io.ReadWriteCloser, error)
}
