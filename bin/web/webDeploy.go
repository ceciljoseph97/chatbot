package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	enableWs   = flag.Bool("enableWs", false, "enable WebSocket endpoint")
	dev        = flag.Bool("dev", false, "developer mode")
	chatbotURL = flag.String("chatbot_url", "http://localhost:9090/get_response", "URL of the chatbot service")
)

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	if *enableWs {
		http.HandleFunc("/ws", handleWebSocket)
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

	response, err := getChatbotResponse(req.Message)
	if err != nil {
		http.Error(w, "Failed to get response from chatbot", http.StatusInternalServerError)
		if *dev {
			log.Printf("Error getting response from chatbot: %v", err)
		}
		return
	}

	resp := struct {
		Reply string `json:"reply"`
	}{
		Reply: response,
	}

	json.NewEncoder(w).Encode(resp)
}

func getChatbotResponse(message string) (string, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"message": message,
	})

	resp, err := http.Post(*chatbotURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chatbot service returned status: %s", resp.Status)
	}

	var res struct {
		Reply string `json:"reply"`
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res.Reply, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}
		clientMessage := string(message)
		if *dev {
			log.Printf("Received message: %s", clientMessage)
		}

		response, err := getChatbotResponse(clientMessage)
		if err != nil {
			log.Println("Error getting response from chatbot:", err)
			err = conn.WriteMessage(websocket.TextMessage, []byte("Error getting response from chatbot"))
			if err != nil {
				log.Println("Write error:", err)
				break
			}
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
