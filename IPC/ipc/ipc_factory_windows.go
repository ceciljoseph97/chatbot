package ipc

func NewIPC(pipeName string) IPC {
	return NewIPCWindows(pipeName)
}
