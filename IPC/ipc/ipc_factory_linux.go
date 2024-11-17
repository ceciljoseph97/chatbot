package ipc

func NewIPC(socketPath string) IPC {
	return NewIPCLinux(socketPath)
}
