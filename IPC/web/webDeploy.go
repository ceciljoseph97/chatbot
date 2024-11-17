package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"golangChatBot/IPC/ipc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Reply     string `json:"reply,omitempty"`
	Error     string `json:"error,omitempty"`
}

var (
	ipcPipeName = flag.String("ipc_pipe", `\\.\pipe\chatbot_pipe`, "Named pipe for IPC communication with Chatbot (Windows) or socket path for Linux")
	enableWs    = flag.Bool("enableWs", false, "Enable WebSocket endpoint")
	devMode     = flag.Bool("dev", false, "Developer mode (verbose logging)")
	listenAddr  = flag.String("listen", ":8080", "Address to listen on for Web Server")

	ipcConn      io.WriteCloser
	ipcConnMutex sync.Mutex

	ipcReader    *bufio.Reader
	ipcReaderMux sync.Mutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	flag.Parse()

	log.Printf("Initializing Web Server...")
	ipcInstance := ipc.NewIPC(*ipcPipeName)
	conn, err := ipcInstance.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Chatbot's IPC: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to Chatbot's IPC: %s", *ipcPipeName)

	ipcConn = conn
	ipcReader = bufio.NewReader(conn)

	go func() {
		http.Handle("/", http.FileServer(http.Dir("./static")))

		if *enableWs {
			http.HandleFunc("/ws", handleWebSocket)
			fmt.Println("WebSocket endpoint /ws is enabled")
		} else {
			fmt.Println("WebSocket endpoint /ws is disabled")
		}

		http.HandleFunc("/chat", chatHandler)

		fmt.Printf("Starting Web Server on %s...\n", *listenAddr)
		if err := http.ListenAndServe(*listenAddr, nil); err != nil {
			log.Fatalf("Failed to start Web Server: %v", err)
		}
	}()

	select {}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Invalid request method"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Bad request"}`, http.StatusBadRequest)
		return
	}

	requestID := uuid.New().String()

	msg := Message{
		RequestID: requestID,
		Message:   req.Message,
	}

	log.Printf("Sending message to Chatbot: %+v", msg)

	reply, err := sendMessageToChatbot(msg)
	if err != nil {
		if *devMode {
			log.Printf("Error getting response from Chatbot: %v", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get response from Chatbot"})
		return
	}

	log.Printf("Received reply from Chatbot: %s", reply)

	resp := struct {
		Reply string `json:"reply"`
	}{
		Reply: reply,
	}

	json.NewEncoder(w).Encode(resp)
}

func sendMessageToChatbot(msg Message) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %v", err)
	}

	ipcConnMutex.Lock()
	if ipcConn == nil {
		ipcConnMutex.Unlock()
		return "", fmt.Errorf("not connected to Chatbot's IPC")
	}
	_, err = ipcConn.Write(append(data, '\n'))
	ipcConnMutex.Unlock()
	if err != nil {
		return "", fmt.Errorf("failed to write to Chatbot's IPC: %v", err)
	}

	log.Printf("Message sent to Chatbot: %s", string(data))

	ipcReaderMux.Lock()
	defer ipcReaderMux.Unlock()
	if ipcReader == nil {
		return "", fmt.Errorf("ipcReader is not initialized")
	}
	line, err := ipcReader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from Chatbot's IPC: %v", err)
	}

	log.Printf("Raw response received: %s", string(line))

	var resp Message
	err = json.Unmarshal(line, &resp)
	if err != nil {
		return "", fmt.Errorf("invalid JSON response from Chatbot: %v", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	}

	return resp.Reply, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket Read error: %v", err)
			}
			break
		}
		clientMessage := string(message)
		if *devMode {
			log.Printf("Received WebSocket message: %s", clientMessage)
		}

		requestID := uuid.New().String()
		msg := Message{
			RequestID: requestID,
			Message:   clientMessage,
		}

		log.Printf("Sending WebSocket message to Chatbot: %+v", msg)

		reply, err := sendMessageToChatbot(msg)
		if err != nil {
			log.Println("Error getting response from Chatbot:", err)
			errMsg := "Error getting response from Chatbot"
			if *devMode {
				errMsg = fmt.Sprintf("Error: %v", err)
			}
			conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
			continue
		}

		log.Printf("Received WebSocket reply from Chatbot: %s", reply)

		err = conn.WriteMessage(websocket.TextMessage, []byte(reply))
		if err != nil {
			log.Println("WebSocket Write error:", err)
			break
		}
	}
}
