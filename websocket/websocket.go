package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/lxzan/gws"
)

const (
// PingInterval = 5 * time.Second
// PingWait     = 10 * time.Second
)

type Handler struct{}

type Incoming struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (h *Handler) OnOpen(conn *gws.Conn) {}

func (h *Handler) OnClose(conn *gws.Conn, err error) {}

func (h *Handler) OnPing(conn *gws.Conn, payload []byte) {}

func (h *Handler) OnPong(conn *gws.Conn, payload []byte) {}

func (h *Handler) OnMessage(connection *gws.Conn, message *gws.Message) {

	defer message.Close()

	var incoming Incoming
	if err := json.Unmarshal(message.Bytes(), &incoming); err != nil {
		connection.WriteMessage(gws.OpcodeText, []byte(`{"error":"invalid_json"}`))
		return
	}

	switch incoming.Type {

	case "get_finalization_proof":
		GetFinalizationProof(incoming.Data, connection)
	case "get_leader_rotation_proof":
		GetLeaderRotationProof(incoming.Data, connection)
	default:
		connection.WriteMessage(gws.OpcodeText, []byte(`{"error":"unknown_type"}`))

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

	wsInterface := globals.CONFIGURATION.WebSocketInterface

	wsPort := globals.CONFIGURATION.WebSocketPort

	address := wsInterface + ":" + strconv.Itoa(wsPort)

	log.Printf("WebSocket server listening on %s\n", address)

	if err := http.ListenAndServe(address, nil); err != nil {

		log.Fatalf("WebSocket server failed: %v", err)

	}

}
