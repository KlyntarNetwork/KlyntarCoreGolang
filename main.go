/**
 *
 *
 *    Developed on Earth,Milky Way(Sagittarius A*) by humanity
 *
 *    Date: ~66.5 ml after Chicxulub
 *
 *    Dev:Vlad Chernenko(@MausClaus)
 *
 *
 *    ⟒10⏚19⎎12⟒33⏃☊0⟒⟒⏚401⎅671⏚⏃23⟒38899⎎⎅387847183☊⎅6⏚8308⏃☊72⎅511⏃⏚
 *
 *
 *
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
)

func main() {

	//_________________________________________________PRINT BANNER & GREETING_______________________________________________

	klyntarBannerPrint()

	//_____________________________________________________CONFIG_PROCESS____________________________________________________

	configsRawJson, readError := os.ReadFile(globals.CHAINDATA_PATH + "/configs.json")

	if readError != nil {

		panic("Error while reading configs: " + readError.Error())

	}

	if err := json.Unmarshal(configsRawJson, &globals.CONFIGURATION); err != nil {

		panic("Error with configs parsing: " + err.Error())

	}

	//_____________________________________________________READ GENESIS______________________________________________________

	genesisRawJson, readError := os.ReadFile(globals.CHAINDATA_PATH + "/genesis.json")

	if readError != nil {

		panic("Error while reading genesis: " + readError.Error())

	}

	if err := json.Unmarshal(genesisRawJson, &globals.GENESIS); err != nil {

		panic("Error with genesis parsing: " + err.Error())

	}

	currentUser, _ := user.Current()

	statsStringToPrint := fmt.Sprintf("System info \x1b[31mgolang:%s \033[36;1m/\x1b[31m os info:%s # %s # cpu:%d \033[36;1m/\x1b[31m runned as:%s\x1b[0m", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), currentUser.Username)

	utils.LogWithTime(statsStringToPrint, utils.CYAN_COLOR)

	go signalHandler()

	// Function that runs the main logic

	RunBlockchain()

}

func klyntarBannerPrint() {

	var finalArt string

	banner, err := os.ReadFile("images/banner.txt")

	if err != nil {
		fmt.Println("Error while reading banner:", err)
		return
	}

	//...and add extra colors & changes)

	finalArt = strings.ReplaceAll(string(banner), "Made on Earth for Universe", "\u001b[37mMade on Earth for Universe\u001b[0m")

	finalArt = strings.ReplaceAll(finalArt, "█", "\u001b[38;5;50m█\x1b[0m")
	finalArt = strings.ReplaceAll(finalArt, "#", "\x1b[31m#\x1b[36m")
	finalArt = strings.ReplaceAll(finalArt, ")", "\u001b[38;5;3m)\x1b[0m")
	finalArt = strings.ReplaceAll(finalArt, "(", "\u001b[38;5;57m(\x1b[0m")
	finalArt = strings.ReplaceAll(finalArt, "|", "\u001b[38;5;87m|\x1b[0m")
	finalArt = strings.ReplaceAll(finalArt, "Follow our Github to build the future", "\u001b[38;5;23mFollow our Github to build the future\x1b[0m")
	finalArt += "\x1b[0m\n"

	fmt.Println(finalArt)

}

// Function to handle Ctrl+C interruptions
func signalHandler() {

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	utils.GracefulShutdown()

}
