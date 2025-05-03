package utils

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/gorilla/websocket"
	"lukechampine.com/blake3"
)

// ANSI escape codes for text colors
const (
	RESET_COLOR       = "\033[0m"
	RED_COLOR         = "\033[31;1m"
	DEEP_GREEN_COLOR  = "\u001b[38;5;23m"
	DEEP_ORANGE_COLOR = "\u001b[38;5;202m"
	GREEN_COLOR       = "\033[32;1m"
	YELLOW_COLOR      = "\033[33m"
	BLUE_COLOR        = "\033[34;1m"
	MAGENTA_COLOR     = "\033[38;5;99m"
	CYAN_COLOR        = "\033[36;1m"
	WHITE_COLOR       = "\033[37;1m"
)

var shutdownOnce sync.Once

func OpenWebsocketConnectionsWithQuorum(quorum []string, wsConnMap map[string]*websocket.Conn, msgHandler func([]byte)) {

	for _, validatorID := range quorum {

		// Skip if connection already exists
		if _, exists := wsConnMap[validatorID]; exists {
			continue
		}

		// Fetch data from LevelDB
		raw, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte(validatorID), nil)

		if err != nil {
			continue
		}

		// Parse JSON into PoolStorage
		var pool structures.PoolStorage

		if err := json.Unmarshal(raw, &pool); err != nil {
			continue
		}

		// Check if the validator is active and has a WebSocket URL
		if !pool.Activated || pool.WssPoolURL == "" {
			continue
		}

		// Dial the WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(pool.WssPoolURL, nil)

		if err != nil {
			continue
		}

		// Store the connection
		wsConnMap[validatorID] = conn

		// Monitor connection in background
		go func(id string, c *websocket.Conn) {
			defer c.Close()
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					delete(wsConnMap, id)
					return
				}

				msgHandler(msg)
			}
		}(validatorID, conn)
	}

}

func CleanupWebsocketConnections(quorum []string, wsConnMap map[string]*websocket.Conn) {

	// Build a set of current quorum IDs for fast lookup
	active := make(map[string]struct{})
	for _, id := range quorum {
		active[id] = struct{}{}
	}

	for id, conn := range wsConnMap {
		if _, ok := active[id]; !ok {
			// Validator is no longer in quorum — close and remove
			conn.Close()
			delete(wsConnMap, id)
		}
	}
}

type QuorumWaiter struct {
	responseCh chan string
	done       chan struct{}
	answered   map[string]bool
	timer      *time.Timer
	mu         sync.Mutex
	buf        []string
}

func NewQuorumWaiter(maxQuorumSize int) *QuorumWaiter {
	return &QuorumWaiter{
		responseCh: make(chan string, maxQuorumSize),
		done:       make(chan struct{}),
		answered:   make(map[string]bool, maxQuorumSize),
		timer:      time.NewTimer(0),
		buf:        make([]string, 0, maxQuorumSize),
	}
}

func (qw *QuorumWaiter) sendMessages(targets []string, msg []byte, wsConnMap map[string]*websocket.Conn) {

	for _, id := range targets {
		conn, ok := wsConnMap[id]
		if !ok {
			continue
		}

		go func(id string, c *websocket.Conn) {
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

			_ = c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			_, _, err := c.ReadMessage()
			if err == nil {
				select {
				case qw.responseCh <- id:
				case <-qw.done:
				}
			}
		}(id, conn)
	}
}

func (qw *QuorumWaiter) SendAndWait(
	ctx context.Context,
	message []byte,
	quorum []string,
	wsConnMap map[string]*websocket.Conn,
	majority int,
) bool {

	// Reset internal state
	qw.mu.Lock()

	for k := range qw.answered {
		delete(qw.answered, k)
	}
	qw.buf = qw.buf[:0]
	qw.mu.Unlock()

	// Reset timer and done channel
	if !qw.timer.Stop() {
		select {
		case <-qw.timer.C:
		default:
		}
	}
	qw.timer.Reset(100 * time.Millisecond)
	qw.done = make(chan struct{}) // reset done channel

	// Fast send
	qw.sendMessages(quorum, message, wsConnMap)

	for {
		select {
		case id := <-qw.responseCh:
			qw.mu.Lock()
			if !qw.answered[id] {
				qw.answered[id] = true
			}
			count := len(qw.answered)
			qw.mu.Unlock()

			if count >= majority {
				close(qw.done)
				return true
			}

		case <-qw.timer.C:
			// Retry missing
			qw.mu.Lock()
			qw.buf = qw.buf[:0]
			for _, id := range quorum {
				if !qw.answered[id] {
					qw.buf = append(qw.buf, id)
				}
			}
			qw.mu.Unlock()

			if len(qw.buf) == 0 {
				return false
			}
			qw.timer.Reset(100 * time.Millisecond)
			qw.sendMessages(qw.buf, message, wsConnMap)

		case <-ctx.Done():
			return false
		}
	}
}

func GracefulShutdown() {

	shutdownOnce.Do(func() {

		LogWithTime("\x1b[31;1mKLYNTAR\x1b[36;1m stop has been initiated.Keep waiting...", CYAN_COLOR)

		LogWithTime("Closing server connections...", CYAN_COLOR)

		LogWithTime("Node was gracefully stopped", CYAN_COLOR)

		os.Exit(0)

	})

}

func LogWithTime(msg, msgColor string) {

	formattedDate := time.Now().Format("02 January 2006 at 03:04:05 PM")

	var prefixColor string

	if os.Getenv("KLY_MODE") == "test" {

		prefixColor = DEEP_ORANGE_COLOR

	} else {

		prefixColor = DEEP_GREEN_COLOR

	}

	fmt.Printf(prefixColor+"[%s]"+MAGENTA_COLOR+"(pid:%d)"+msgColor+"  %s\n"+RESET_COLOR, formattedDate, os.Getpid(), msg)

}

func Blake3(data string) string {

	blake3Hash := blake3.Sum256([]byte(data))

	return hex.EncodeToString(blake3Hash[:])

}

func GetUTCTimestampInMilliSeconds() int64 {

	return time.Now().UTC().UnixMilli()

}

type CurrentLeaderData struct {
	IsMeLeader bool
	Url        string
}

func getUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func IsMyCoreVersionOld(thread *structures.ApprovementThread) bool {

	return thread.CoreMajorVersion > globals.CORE_MAJOR_VERSION

}

func EpochStillFresh(thread *structures.ApprovementThread) bool {

	return (thread.EpochHandler.StartTimestamp + uint64(thread.NetworkParameters.EpochTime)) > uint64(getUtcTimestamp())

}

func GetCurrentLeader() CurrentLeaderData {

	globals.APPROVEMENT_THREAD.RWMutex.RLock()

	defer globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

	currentLeaderPubKey := globals.APPROVEMENT_THREAD.Thread.EpochHandler.LeadersSequence[globals.APPROVEMENT_THREAD.Thread.EpochHandler.CurrentLeaderIndex]

	if currentLeaderPubKey == globals.CONFIGURATION.PublicKey {

		return CurrentLeaderData{IsMeLeader: true, Url: ""}

	}

	return CurrentLeaderData{IsMeLeader: false, Url: ""}
}

func IntToBytes(n int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	return buf
}

func BytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
