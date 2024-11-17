package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"golangChatBot/web/perichatbot"
)

var (
	configFile = flag.String("config", "./config_local_gen.yaml", "path to the config file")
	dev        = flag.Bool("dev", false, "developer mode")
	storeFile  = flag.String("c", "PMFuncOverView.gob", "the file to store corpora")
	tops       = flag.Int("t", 1, "the number of answers to return")
	enableWs   = flag.Bool("enableWs", false, "enable WebSocket endpoint")
)
var chatbot *perichatbot.Chatbot

func main() {
	flag.Parse()

	var err error
	chatbot, err = perichatbot.NewChatbot(*configFile, *dev, *storeFile, *tops)
	if err != nil {
		log.Fatalf("Error initializing chatbot: %v", err)
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))

	if *enableWs {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			chatbot.HandleWebSocket(w, r)
		})
		fmt.Println("WebSocket endpoint /ws is enabled")
	} else {
		fmt.Println("WebSocket endpoint /ws is disabled")
	}

	http.HandleFunc("/chat", chatHandler)

	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	response := chatbot.GetResponse(req.Message)

	resp := struct {
		Reply string `json:"reply"`
	}{
		Reply: response,
	}

	json.NewEncoder(w).Encode(resp)
}
