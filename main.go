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
	"os"
	"os/user"
	"runtime"
	"strings"

	klyUtils "github.com/KLYN74R/KlyntarCoreGolang/KLY_Utils"

	tachyon "github.com/KLYN74R/KlyntarCoreGolang/KLY_Workflows/dev_tachyon"
)

/*
****************************************************************************************************************
*                                                                                                              *
*                                                                                                              *
*                                    ░██████╗████████╗░█████╗░██████╗░████████╗                                *
*                                    ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗╚══██╔══╝                                *
*                                    ╚█████╗░░░░██║░░░███████║██████╔╝░░░██║░░░                                *
*                                    ░╚═══██╗░░░██║░░░██╔══██║██╔══██╗░░░██║░░░                                *
*                                    ██████╔╝░░░██║░░░██║░░██║██║░░██║░░░██║░░░                                *
*                                    ╚═════╝░░░░╚═╝░░░╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░                                *
*                                                                                                              *
*                                                                                                              *
****************************************************************************************************************
 */

//_____________________________________________________DEFINE GLOBAL ACCESS VALUES____________________________________________________

// Pathes to 3 main direcories
var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string

// Global configs (resolved by <CONFIGS_PATH>, example available in KLY_Workflows/dev_tachyon/templates/configs.json)
var CONFIGS map[string]interface{}

// Load genesis from JSON file to pre-set the state
var GENESIS map[string]interface{}

func main() {

	//_________________________________________________PRINT BANNER & GREETING_______________________________________________

	KlyntarBannerPrint()

	PrepareRequiredPath()

	//_____________________________________________________CONFIG_PROCESS____________________________________________________

	configsRawJson, readError := os.ReadFile(CONFIGS_PATH)

	if readError != nil {

		panic("Error while reading configs: " + readError.Error())

	}

	if err := json.Unmarshal(configsRawJson, &CONFIGS); err != nil {

		panic("Error with configs parsing: " + err.Error())

	}

	//_____________________________________________________READ GENESIS______________________________________________________

	genesisRawJson, readError := os.ReadFile(GENESIS_PATH)

	if readError != nil {

		panic("Error while reading genesis: " + readError.Error())

	}

	if err := json.Unmarshal(genesisRawJson, &GENESIS); err != nil {

		panic("Error with genesis parsing: " + err.Error())

	}

	//_________________________________________PREPARE DIRECTORIES FOR CHAINDATA_____________________________________________

	// Check if exists
	if _, err := os.Stat(CHAINDATA_PATH); os.IsNotExist(err) {

		// If no - create
		if err := os.MkdirAll(CHAINDATA_PATH, os.ModePerm); err != nil {

			panic("Error with creating directory for chaindata: " + err.Error())

		}

	}

	currentUser, _ := user.Current()

	statsStringToPrint := fmt.Sprintf("System info \x1b[31mgolang:%s \033[36;1m/\x1b[31m os info:%s # %s # cpu:%d \033[36;1m/\x1b[31m runned as:%s\x1b[0m", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), currentUser.Username)

	klyUtils.LogWithTime(statsStringToPrint, klyUtils.CYAN_COLOR)

	tachyon.RunBlockchain()

}

/*



                                .do-"""""'-o..
                             .o""            ""..
                           ,,''                 ``b.
                          d'                      ``b
                         d`d:                       `b.
                        ,,dP                         `Y.
                       d`88                           `8.
 ooooooooooooooooood888`88'                            `88888888888bo,
d"""    `""""""""""""Y:d8P                              8,          `b
8                    P,88b                             ,`8           8
8                   ::d888,                           ,8:8.          8                              ██████╗ ███████╗██╗   ██╗███████╗██╗      ██████╗ ██████╗ ███████╗██████╗
:                   dY88888                           `' ::          8                              ██╔══██╗██╔════╝██║   ██║██╔════╝██║     ██╔═══██╗██╔══██╗██╔════╝██╔══██╗
:                   8:8888                               `b          8                              ██║  ██║█████╗  ██║   ██║█████╗  ██║     ██║   ██║██████╔╝█████╗  ██║  ██║
:                   Pd88P',...                     ,d888o.8          8                              ██║  ██║██╔══╝  ╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██╔═══╝ ██╔══╝  ██║  ██║
:                   :88'dd888888o.                d8888`88:          8                              ██████╔╝███████╗ ╚████╔╝ ███████╗███████╗╚██████╔╝██║     ███████╗██████╔╝
:                  ,:Y:d8888888888b             ,d88888:88:          8                              ╚═════╝ ╚══════╝  ╚═══╝  ╚══════╝╚══════╝ ╚═════╝ ╚═╝     ╚══════╝╚═════╝
:                  :::b88d888888888b.          ,d888888bY8b          8
                    b:P8;888888888888.        ,88888888888P          8
                    8:b88888888888888:        888888888888'          8
                    8:8.8888888888888:        Y8888888888P           8                              ███████╗ ██████╗ ██████╗     ██████╗ ███████╗ ██████╗ ██████╗ ██╗     ███████╗
,                   YP88d8888888888P'          ""888888"Y            8                              ██╔════╝██╔═══██╗██╔══██╗    ██╔══██╗██╔════╝██╔═══██╗██╔══██╗██║     ██╔════╝
:                   :bY8888P"""""''                     :            8                              █████╗  ██║   ██║██████╔╝    ██████╔╝█████╗  ██║   ██║██████╔╝██║     █████╗
:                    8'8888'                            d            8                              ██╔══╝  ██║   ██║██╔══██╗    ██╔═══╝ ██╔══╝  ██║   ██║██╔═══╝ ██║     ██╔══╝
:                    :bY888,                           ,P            8                              ██║     ╚██████╔╝██║  ██║    ██║     ███████╗╚██████╔╝██║     ███████╗███████╗
:                     Y,8888           d.  ,-         ,8'            8                              ╚═╝      ╚═════╝ ╚═╝  ╚═╝    ╚═╝     ╚══════╝ ╚═════╝ ╚═╝     ╚══════╝╚══════╝
:                     `8)888:           '            ,P'             8
:                      `88888.          ,...        ,P               8
:                       `Y8888,       ,888888o     ,P                8                              ██████╗ ██╗   ██╗    ██╗  ██╗██╗  ██╗   ██╗███╗   ██╗████████╗ █████╗ ██████╗     ████████╗███████╗ █████╗ ███╗   ███╗
:                         Y888b      ,88888888    ,P'                8                              ██╔══██╗╚██╗ ██╔╝    ██║ ██╔╝██║  ╚██╗ ██╔╝████╗  ██║╚══██╔══╝██╔══██╗██╔══██╗    ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║
:                          `888b    ,888888888   ,,'                 8                              ██████╔╝ ╚████╔╝     █████╔╝ ██║   ╚████╔╝ ██╔██╗ ██║   ██║   ███████║██████╔╝       ██║   █████╗  ███████║██╔████╔██║
:                           `Y88b  dPY888888OP   :'                  8                              ██╔══██╗  ╚██╔╝      ██╔═██╗ ██║    ╚██╔╝  ██║╚██╗██║   ██║   ██╔══██║██╔══██╗       ██║   ██╔══╝  ██╔══██║██║╚██╔╝██║
:                             :88.,'.   `' `8P-"b.                   8                              ██████╔╝   ██║       ██║  ██╗███████╗██║   ██║ ╚████║   ██║   ██║  ██║██║  ██║       ██║   ███████╗██║  ██║██║ ╚═╝ ██║
:.                             )8P,   ,b '  -   ``b                  8                              ╚═════╝    ╚═╝       ╚═╝  ╚═╝╚══════╝╚═╝   ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝       ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝
::                            :':   d,'d`b, .  - ,db                 8
::                            `b. dP' d8':      d88'                 8
::                             '8P" d8P' 8 -  d88P'                  8
::                            d,' ,d8'  ''  dd88'                    8
::                           d'   8P'  d' dd88'8                     8
 :                          ,:   `'   d:ddO8P' `b.                   8
 :                  ,dooood88: ,    ,d8888""    ```b.                8
 :               .o8"'""""""Y8.b    8 `"''    .o'  `"""ob.           8
 :              dP'         `8:     K       dP''        "`Yo.        8
 :             dP            88     8b.   ,d'              ``b       8
 :             8.            8P     8""'  `"                 :.      8                              ██╗   ██╗   ██████╗
 :            :8:           :8'    ,:                        ::      8                              ██║   ██║  ██╔════╝
 :            :8:           d:    d'                         ::      8                              ██║   ██║  ██║
 :            :8:          dP   ,,'                          ::      8                              ╚██╗ ██╔╝  ██║
 :            `8:     :b  dP   ,,                            ::      8                               ╚████╔╝██╗╚██████╗██╗
 :            ,8b     :8 dP   ,,                             d       8                                ╚═══╝ ╚═╝ ╚═════╝╚═╝
 :            :8P     :8dP    d'                       d     8       8
 :            :8:     d8P    d'                      d88    :P       8
 :            d8'    ,88'   ,P                     ,d888    d'       8
 :            88     dP'   ,P                      d8888b   8        8
 '           ,8:   ,dP'    8.                     d8''88'  :8        8
             :8   d8P'    d88b                   d"'  88   :8        8
             d: ,d8P'    ,8P""".                      88   :P        8
             8 ,88P'     d'                           88   ::        8
            ,8 d8P       8                            88   ::        8
            d: 8P       ,:  -hrr-                    :88   ::        8
            8',8:,d     d'                           :8:   ::        8
           ,8,8P'8'    ,8                            :8'   ::        8
           :8`' d'     d'                            :8    ::        8
           `8  ,P     :8                             :8:   ::        8
            8, `      d8.                            :8:   8:        8
            :8       d88:                            d8:   8         8
 ,          `8,     d8888                            88b   8         8
 :           88   ,d::888                            888   Y:        8
 :           YK,oo8P :888                            888.  `b        8
 :           `8888P  :888:                          ,888:   Y,       8
 :            ``'"   `888b                          :888:   `b       8
 :                    8888                           888:    ::      8
 :                    8888:                          888b     Y.     8,
 :                    8888b                          :888     `b     8:
 :                    88888.                         `888,     Y     8:
 ``ob...............--"""""'----------------------`""""""""'"""`'"""""

*/

func KlyntarBannerPrint() {

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

// Function to resolve the path to 3 main directories - CHAINDATA, GENESIS, CONFIGS
func PrepareRequiredPath() {

	if os.Getenv("CHANDATA_PATH") == "" {

		CHAINDATA_PATH = "CHAINDATA"

	} else {

		CHAINDATA_PATH = os.Getenv("CHANDATA_PATH")

	}

	if os.Getenv("GENESIS_PATH") == "" {

		GENESIS_PATH = "GENESIS"

	} else {

		GENESIS_PATH = os.Getenv("GENESIS_PATH")

	}

	if os.Getenv("CONFIGS_PATH") == "" {

		CONFIGS_PATH = "CONFIGS"

	} else {

		CONFIGS_PATH = os.Getenv("CONFIGS_PATH")

	}

}
