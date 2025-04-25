/**
 *
 *
 *
 *
 *
 *                                                               ██╗  ██╗██╗  ██╗   ██╗███╗   ██╗████████╗ █████╗ ██████╗
 *                                                               ██║ ██╔╝██║  ╚██╗ ██╔╝████╗  ██║╚══██╔══╝██╔══██╗██╔══██╗
 *                                                               █████╔╝ ██║   ╚████╔╝ ██╔██╗ ██║   ██║   ███████║██████╔╝
 *                                                               ██╔═██╗ ██║    ╚██╔╝  ██║╚██╗██║   ██║   ██╔══██║██╔══██╗
 *                                                               ██║  ██╗███████╗██║   ██║ ╚████║   ██║   ██║  ██║██║  ██║
 *                                                               ╚═╝  ╚═╝╚══════╝╚═╝   ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝
 *
 *
 *
 *                                                               Developed on Earth,Milky Way(Sagittarius A*) by humanity
 *
 *
 *                                                                          Date: ~66.5 ml after Chicxulub
 *
 *
 *                                                                          Dev:Vlad Chernenko(@V14D4RT3M)
 *
 *
 *                                                       ⟒10⏚19⎎12⟒33⏃☊0⟒⟒⏚401⎅671⏚⏃23⟒38899⎎⎅387847183☊⎅6⏚8308⏃☊72⎅511⏃⏚
 *
 *
 *
 *
 *
 *
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
)

func main() {

	//_________________________________________________PRINT BANNER & GREETING_______________________________________________

	klyntarBannerPrint()

	prepareRequiredPaths()

	//_____________________________________________________CONFIG_PROCESS____________________________________________________

	configsRawJson, readError := os.ReadFile(globals.CONFIGS_PATH + "/configs.json")

	if readError != nil {

		panic("Error while reading configs: " + readError.Error())

	}

	if err := json.Unmarshal(configsRawJson, &globals.CONFIGURATION); err != nil {

		panic("Error with configs parsing: " + err.Error())

	}

	//_____________________________________________________READ GENESIS______________________________________________________

	genesisRawJson, readError := os.ReadFile(globals.GENESIS_PATH + "/genesis.json")

	if readError != nil {

		panic("Error while reading genesis: " + readError.Error())

	}

	if err := json.Unmarshal(genesisRawJson, &globals.GENESIS); err != nil {

		panic("Error with genesis parsing: " + err.Error())

	}

	//_________________________________________PREPARE DIRECTORIES FOR CHAINDATA_____________________________________________

	// Check if exists
	if _, err := os.Stat(globals.CHAINDATA_PATH); os.IsNotExist(err) {

		// If no - create
		if err := os.MkdirAll(globals.CHAINDATA_PATH, os.ModePerm); err != nil {

			panic("Error with creating directory for chaindata: " + err.Error())

		}

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

	if os.Getenv("KLY_MODE") == "main" {

		// Read banner

		banner, err := os.ReadFile("images/banner.txt")

		if err != nil {
			fmt.Println("Error while reading banner:", err)
			return
		}

		//...and add extra colors & changes)

		finalArt = strings.ReplaceAll(string(banner), "Made on Earth for Universe", "\x1b[31mMade on Earth for Universe\x1b[36m")

		finalArt = strings.ReplaceAll(finalArt, "REMEMBER:To infinity and beyond!", "\x1b[31mREMEMBER:To infinity and beyond!\u001b[37m")
		finalArt = strings.ReplaceAll(finalArt, "≈", "\x1b[31m≈\x1b[36m")
		finalArt = strings.ReplaceAll(finalArt, "#", "\x1b[31m#\u001b[37m")

	} else {

		testmodeBanner, err := os.ReadFile("images/testmode_banner.txt")

		if err != nil {
			fmt.Println("Error while reading banner:", err)
			return
		}

		//...and add extra colors & changes)

		finalArt = strings.ReplaceAll(string(testmodeBanner), "Made on Earth for Universe", "\u001b[38;5;87mMade on Earth for Universe\u001b[37m")

		finalArt = strings.ReplaceAll(finalArt, "REMEMBER:To infinity and beyond!", "\u001b[38;5;87mREMEMBER:To infinity and beyond!\u001b[37m")
		finalArt = strings.ReplaceAll(finalArt, "≈", "\x1b[31m≈\u001b[37m")

		finalArt = strings.ReplaceAll(finalArt, "█", "\u001b[38;5;202m█\u001b[37m")
		finalArt = strings.ReplaceAll(finalArt, "=", "\u001b[38;5;87m═\u001b[37m")
		finalArt = strings.ReplaceAll(finalArt, "╝", "\u001b[38;5;87m╝\u001b[37m")
		finalArt = strings.ReplaceAll(finalArt, "╚", "\u001b[38;5;87m╚\u001b[37m")

		finalArt = strings.ReplaceAll(finalArt, "#", "\u001b[38;5;202m#\u001b[37m")

	}

	fmt.Println(finalArt)

}

// Function to resolve the paths to 3 main directories - CHAINDATA, GENESIS, CONFIGS
func prepareRequiredPaths() {

	baseDir := os.Getenv("SYMBIOTE_DIR")

	if baseDir == "" {

		log.Fatal("SYMBIOTE_DIR environment variable is not set")

	}

	baseDir = strings.TrimRight(baseDir, "/")

	if !filepath.IsAbs(baseDir) {

		log.Fatalf("SYMBIOTE_DIR must be an absolute path, got: %s", baseDir)

	}

	globals.CHAINDATA_PATH = baseDir + "/CHAINDATA"

	globals.GENESIS_PATH = baseDir + "/GENESIS"

	globals.CONFIGS_PATH = baseDir + "/CONFIGS"

}

// Function to handle Ctrl+C interruptions
func signalHandler() {

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	utils.GracefulShutdown()

}
