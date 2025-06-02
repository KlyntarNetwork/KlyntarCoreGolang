package globals

import (
	"os"
	"strconv"
	"sync"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/syndtr/goleveldb/leveldb"
)

var CORE_MAJOR_VERSION int = func() int {

	data, err := os.ReadFile("version.txt")

	if err != nil {
		panic("Failed to read version.txt: " + err.Error())
	}

	version, err := strconv.Atoi(string(data))

	if err != nil {
		panic("Invalid version format: " + err.Error())
	}

	return version

}()

var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string // pathes to 3 main directories

var CONFIGURATION structures.NodeLevelConfig

var GENESIS structures.Genesis

var MEMPOOL struct {
	Slice []structures.Transaction
	Mutex sync.Mutex
}

var GENERATION_THREAD_METADATA_HANDLER structures.GenerationThreadMetadataHandler

var APPROVEMENT_THREAD_METADATA_HANDLER = struct {
	RWMutex sync.RWMutex
	Handler structures.ApprovementThreadMetadataHandler
}{
	Handler: structures.ApprovementThreadMetadataHandler{
		CoreMajorVersion: -1,
		Cache:            make(map[string]*structures.PoolStorage),
	},
}

var BLOCKS, EPOCH_DATA, APPROVEMENT_THREAD_METADATA, FINALIZATION_VOTING_STATS *leveldb.DB
