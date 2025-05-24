package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/websocket"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	prepareBlockchain()

	//_________________________ RUN SEVERAL THREADS _________________________

	// //✅1.Thread to find AEFPs and change the epoch for AT
	// go life.EpochRotationThread()

	// //✅2.Share our blocks within quorum members and get the finalization proofs
	// go life.BlocksSharingAndProofsGrabingThread()

	// //✅3.Thread to propose AEFPs to move to next epoch
	// go life.NewEpochProposerThread()

	// //✅4.Start to generate blocks
	// go life.BlocksGenerationThread()

	// //✅5.Start a separate thread to work with voting for blocks in a sync way (for security)
	// go life.LeaderRotationThread()

	serverAddr := globals.CONFIGURATION.Interface + ":" + strconv.Itoa(globals.CONFIGURATION.Port)

	err := fasthttp.ListenAndServe(serverAddr, NewRouter())

	if err != nil {
		log.Fatalf("Error in server: %s", err)
	}

	websocket.CreateWebsocketServer()

}

func prepareBlockchain() {

	// Create dir
	if _, err := os.Stat(globals.CHAINDATA_PATH); os.IsNotExist(err) {
		if err := os.MkdirAll(globals.CHAINDATA_PATH, 0755); err != nil {
			return
		}
	}

	// Load GT
	if data, err := globals.BLOCKS.Get([]byte("GT"), nil); err == nil && data != nil {
		var gt structures.GenerationThread
		if err := json.Unmarshal(data, &gt); err == nil {
			globals.GENERATION_THREAD = gt
		} else {

			fmt.Println("failed to unmarshal GENERATION_THREAD: %w", err)

			return

		}
	}

	// Load AT
	if data, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte("AT"), nil); err == nil && data != nil {
		var at structures.ApprovementThread
		if err := json.Unmarshal(data, &at); err == nil {
			globals.APPROVEMENT_THREAD.Thread = at
		} else {

			fmt.Println("failed to unmarshal APPROVEMENT_THREAD: %w", err)

			return

		}
	}

	// Init genesis if version is -1
	if globals.APPROVEMENT_THREAD.Thread.CoreMajorVersion == -1 {

		setGenesisToState()

		serialized, err := json.Marshal(globals.APPROVEMENT_THREAD.Thread)
		if err != nil {
			fmt.Println("failed to marshal APPROVEMENT_THREAD: %w", err)
			return
		}

		if err := globals.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), serialized, nil); err != nil {
			fmt.Println("failed to save APPROVEMENT_THREAD: %w", err)
			return
		}

		return
	}

	// Version check
	if utils.IsMyCoreVersionOld(&globals.APPROVEMENT_THREAD.Thread) {

		utils.LogWithTime("New version detected on APPROVEMENT_THREAD. Please, upgrade your node software", utils.YELLOW_COLOR)

		if data, err := os.ReadFile("images/events/update.txt"); err == nil {
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

		poolStorage.Activated = true

		serialized, err := json.Marshal(poolStorage)

		if err != nil {
			return err
		}

		batch.Put([]byte(poolPubKey+"(POOL)_STORAGE_POOL"), serialized)

		poolsRegistryForEpochHandler[poolPubKey] = struct{}{}

	}

	globals.APPROVEMENT_THREAD.Thread.CoreMajorVersion = globals.GENESIS.CoreMajorVersion

	globals.APPROVEMENT_THREAD.Thread.NetworkParameters = globals.GENESIS.NetworkParameters

	// Write batch
	if err := globals.APPROVEMENT_THREAD_METADATA.Write(batch, nil); err != nil {
		return err
	}

	hashInput := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" + globals.GENESIS.NetworkID

	initEpochHash := utils.Blake3(hashInput)

	// Create new epoch handler
	epoch := structures.EpochHandler{
		Id:                 0,
		Hash:               initEpochHash,
		PoolsRegistry:      poolsRegistryForEpochHandler,
		ShardsRegistry:     shardsRegistry,
		StartTimestamp:     epochTimestamp,
		Quorum:             []string{}, // will be assigned
		LeadersSequence:    []string{}, // will be assigned
		CurrentLeaderIndex: 0,
	}

	// Assign quorum
	epoch.Quorum = common_functions.GetCurrentEpochQuorum(&epoch, globals.APPROVEMENT_THREAD.Thread.NetworkParameters.QuorumSize, initEpochHash)

	// Assign sequence of leaders
	common_functions.SetLeadersSequence(&epoch, initEpochHash)

	globals.APPROVEMENT_THREAD.Thread.EpochHandler = epoch

	return nil

}
