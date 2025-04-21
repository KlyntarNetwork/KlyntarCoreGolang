package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lxzan/gws"
)

const (
// PingInterval = 5 * time.Second
// PingWait     = 10 * time.Second
)

type Handler struct{}

func (h *Handler) OnOpen(conn *gws.Conn) {}

func (h *Handler) OnClose(conn *gws.Conn, err error) {}

func (h *Handler) OnPing(conn *gws.Conn, payload []byte) {}

func (h *Handler) OnPong(conn *gws.Conn, payload []byte) {}

func (h *Handler) OnMessage(conn *gws.Conn, message *gws.Message) {

	defer message.Close()

	type Incoming struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}

	var incoming Incoming
	if err := json.Unmarshal(message.Bytes(), &incoming); err != nil {
		conn.WriteMessage(gws.OpcodeText, []byte(`{"error":"invalid_json"}`))
		return
	}

	switch incoming.Type {
	case "ping":
		conn.WriteMessage(gws.OpcodeText, []byte(`{"type":"pong"}`))
	case "echo":
		response, _ := json.Marshal(map[string]any{
			"type": "echo",
			"data": incoming.Data,
		})
		conn.WriteMessage(gws.OpcodeText, response)
	case "broadcast":
		log.Println("Broadcast message:", incoming.Data)
	default:
		conn.WriteMessage(gws.OpcodeText, []byte(`{"error":"unknown_type"}`))
	}
}

func CreateWebsocketServer() {

	upgrader := gws.NewUpgrader(&Handler{}, &gws.ServerOption{
		ParallelEnabled:   true,
		Recovery:          gws.Recovery,
		PermessageDeflate: gws.PermessageDeflate{Enabled: true},
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		go func() {
			conn.ReadLoop()
		}()
	})

	address := "0.0.0.0:6666"
	log.Printf("WebSocket server listening on %s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("WebSocket server failed: %v", err)
	}
}
