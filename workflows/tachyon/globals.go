package tachyon

import (
	"os"
	"strconv"
	"sync"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
	"github.com/syndtr/goleveldb/leveldb"
)

func GetCoreMajorVersion(versionFilePath string) (uint, error) {

	versionData, err := os.ReadFile(versionFilePath)
	if err != nil {
		return 0, err
	}

	majorVersion, err := strconv.ParseUint(string(versionData), 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(majorVersion), nil
}

var CORE_MAJOR_VERSION uint = func() uint {

	version, err := GetCoreMajorVersion("version.txt")

	if err != nil {
		panic("Failed to get core version: " + err.Error())
	}

	return version

}()

// Pathes to 3 main direcories

var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string

var CONFIGS, GENESIS map[string]interface{}

var MEMPOOL struct {
	slice []Transaction
	mutex sync.RWMutex
}

var APPROVEMENT_THREAD_CACHE = make(map[string]interface{})

var GENERATION_THREAD threads.GenerationThread

var APPROVEMENT_THREAD threads.ApprovementThread

var BLOCKS, EPOCH_DATA, APPROVEMENT_THREAD_METADATA, FINALIZATION_VOTING_STATS *leveldb.DB = nil, nil, nil, nil

var VOTING_REQUESTS chan struct{}
