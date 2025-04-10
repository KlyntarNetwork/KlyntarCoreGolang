package tachyon

import (
	"os"
	"strconv"

	threads "github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
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

//_____________________________________________________DEFINE GLOBAL ACCESS VALUES____________________________________________________

var CORE_MAJOR_VERSION uint

// Pathes to 3 main direcories
var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string

// Global configs (resolved by <CONFIGS_PATH>, example available in workflows/tachyon/templates/configs.json)
var CONFIGS map[string]interface{}

// Load genesis from JSON file to pre-set the state
var GENESIS map[string]interface{}

var GLOBAL_CACHES = struct{}{}

var GENERATION_THREAD threads.GenerationThread

var APPROVEMENT_THREAD threads.ApprovementThread

var BLOCKCHAIN_DATABASES = struct {
	BLOCKS, EPOCH_DATA, APPROVEMENT_THREAD_METADATA, FINALIZATION_VOTING_STATS *leveldb.DB
}{
	nil, nil, nil, nil,
}
