package utils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
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

	currentLeaderPubKey := globals.APPROVEMENT_THREAD.EpochHandler.LeadersSequence[globals.APPROVEMENT_THREAD.EpochHandler.CurrentLeaderIndex]

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
