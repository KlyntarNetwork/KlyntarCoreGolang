package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/life"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/websocket"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	prepareBlockchain()

	//_________________________ RUN SEVERAL THREADS _________________________

	//✅1.Thread to find AEFPs and change the epoch for AT
	go life.EpochRotationThread()

	//✅2.Share our blocks within quorum members and get the finalization proofs
	go life.BlocksSharingAndProofsGrabingThread()

	//✅3.Thread to propose AEFPs to move to next epoch
	go life.NewEpochProposerThread()

	//✅4.Start to generate blocks
	go life.BlocksGenerationThread()

	//✅5.Start a separate thread to work with voting for blocks in a sync way (for security)
	go life.LeaderRotationThread()

	//___________________ RUN SERVERS - WEBSOCKET AND HTTP __________________

	go websocket.CreateWebsocketServer()

	serverAddr := globals.CONFIGURATION.Interface + ":" + strconv.Itoa(globals.CONFIGURATION.Port)

	utils.LogWithTime(fmt.Sprintf("Server is starting at http://%s ...✅", serverAddr), utils.CYAN_COLOR)

	err := fasthttp.ListenAndServe(serverAddr, NewRouter())

	if err != nil {

		utils.LogWithTime(fmt.Sprintf("Error in server: %s", err), utils.RED_COLOR)

	}

}

func prepareBlockchain() {

	// Create dir for chaindata
	if _, err := os.Stat(globals.CHAINDATA_PATH); os.IsNotExist(err) {

		if err := os.MkdirAll(globals.CHAINDATA_PATH, 0755); err != nil {

			return

		}

	}

	globals.BLOCKS = utils.OpenDb("BLOCKS")
	globals.EPOCH_DATA = utils.OpenDb("EPOCH_DATA")
	globals.APPROVEMENT_THREAD_METADATA = utils.OpenDb("APPROVEMENT_THREAD_METADATA")
	globals.FINALIZATION_VOTING_STATS = utils.OpenDb("FINALIZATION_VOTING_STATS")

	// Load GT - Generation Thread handler
	if data, err := globals.BLOCKS.Get([]byte("GT"), nil); err == nil {

		var gtHandler structures.GenerationThreadMetadataHandler

		if err := json.Unmarshal(data, &gtHandler); err == nil {

			globals.GENERATION_THREAD_METADATA_HANDLER = gtHandler

		} else {

			fmt.Printf("failed to unmarshal GENERATION_THREAD: %v\n", err)

			return

		}
	} else {

		// Create initial generation thread handler

		globals.GENERATION_THREAD_METADATA_HANDLER = structures.GenerationThreadMetadataHandler{

			EpochFullId: utils.Blake3("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+globals.GENESIS.NetworkId) + "#-1",
			PrevHash:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			NextIndex:   0,
		}

	}

	// Load AT - Approvement Thread handler

	if data, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte("AT"), nil); err == nil {

		var atHandler structures.ApprovementThreadMetadataHandler

		if err := json.Unmarshal(data, &atHandler); err == nil {

			if atHandler.Cache == nil {

				atHandler.Cache = make(map[string]*structures.PoolStorage)

			}

			globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler = atHandler

		} else {

			fmt.Printf("failed to unmarshal APPROVEMENT_THREAD: %v\n", err)

			return

		}

	}

	// Init genesis if version is -1
	if globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.CoreMajorVersion == -1 {

		setGenesisToState()

		serialized, err := json.Marshal(globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler)

		if err != nil {

			fmt.Printf("failed to marshal APPROVEMENT_THREAD: %v\n", err)

			return

		}

		if err := globals.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), serialized, nil); err != nil {

			fmt.Printf("failed to save APPROVEMENT_THREAD: %v\n", err)

			return

		}

		return
	}

	// Version check
	if utils.IsMyCoreVersionOld(&globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler) {

		utils.LogWithTime("New version detected on APPROVEMENT_THREAD. Please, upgrade your node software", utils.YELLOW_COLOR)

		if data, err := os.ReadFile("images/update.txt"); err == nil {

			fmt.Println(string(data))

		}

		utils.GracefulShutdown()

	}

}

func setGenesisToState() error {

	batch := new(leveldb.Batch)

	epochTimestamp := globals.GENESIS.FirstEpochStartTimestamp

	poolsRegistryForEpochHandler := make(map[string]struct{})

	shardsRegistry := []string{globals.GENESIS.Shard}

	// __________________________________ Load info about pools __________________________________

	for poolPubKey, poolStorage := range globals.GENESIS.Pools {

		serialized, err := json.Marshal(poolStorage)

		if err != nil {
			return err
		}

		batch.Put([]byte(poolPubKey+"(POOL)_STORAGE_POOL"), serialized)

		poolsRegistryForEpochHandler[poolPubKey] = struct{}{}

	}

	globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.CoreMajorVersion = globals.GENESIS.CoreMajorVersion

	globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters = globals.GENESIS.NetworkParameters

	// Commit changes
	if err := globals.APPROVEMENT_THREAD_METADATA.Write(batch, nil); err != nil {
		return err
	}

	hashInput := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" + globals.GENESIS.NetworkId

	initEpochHash := utils.Blake3(hashInput)

	// Create new epochHandler handler
	epochHandler := structures.EpochDataHandler{
		Id:                 0,
		Hash:               initEpochHash,
		PoolsRegistry:      poolsRegistryForEpochHandler,
		ShardsRegistry:     shardsRegistry,
		StartTimestamp:     epochTimestamp,
		Quorum:             []string{}, // will be assigned
		LeadersSequence:    []string{}, // will be assigned
		CurrentLeaderIndex: 0,
	}

	// Assign quorum - pseudorandomly and in deterministic way
	epochHandler.Quorum = common_functions.GetCurrentEpochQuorum(&epochHandler, globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters.QuorumSize, initEpochHash)

	// Now set the block generators for epoch pseudorandomly and in deterministic way
	common_functions.SetLeadersSequence(&epochHandler, initEpochHash)

	globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler = epochHandler

	return nil

}
