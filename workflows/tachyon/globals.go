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

var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string // pathes to 3 main directories

var CONFIGS, GENESIS map[string]any

var MEMPOOL struct {
	slice []Transaction
	mutex sync.RWMutex
}

var APPROVEMENT_THREAD_CACHE = make(map[string]any)

var FINALIZATION_PROOFS_CACHE = make(map[string]map[string]string)

var TEMP_CACHE = make(map[string]any)

var GENERATION_THREAD threads.GenerationThread

var APPROVEMENT_THREAD threads.ApprovementThread

var BLOCKS, EPOCH_DATA, APPROVEMENT_THREAD_METADATA, FINALIZATION_VOTING_STATS *leveldb.DB

var VOTING_REQUESTS chan struct{}
