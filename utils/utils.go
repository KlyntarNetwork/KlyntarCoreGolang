package utils

import (
	"fmt"
	"os"
	"time"

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

	return fmt.Sprintf("%x", blake3Hash)

}

func GetUTCTimestampInMilliSeconds() int64 {

	return time.Now().UTC().UnixMilli()

}
