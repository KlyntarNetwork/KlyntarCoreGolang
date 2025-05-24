package globals

import (
	"os"
	"strconv"
	"sync"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/syndtr/goleveldb/leveldb"
)

func getCoreMajorVersion(versionFilePath string) (int, error) {

	versionData, err := os.ReadFile(versionFilePath)
	if err != nil {
		return 0, err
	}

	majorVersion, err := strconv.ParseInt(string(versionData), 10, 64)
	if err != nil {
		return 0, err
	}

	return int(majorVersion), nil
}

var CORE_MAJOR_VERSION int = func() int {

	version, err := getCoreMajorVersion("version.txt")

	if err != nil {
		panic("Failed to get core version: " + err.Error())
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

var GENERATION_THREAD_HANDLER structures.GenerationThread

var APPROVEMENT_THREAD_HANDLER = struct {
	RWMutex sync.RWMutex
	Thread  structures.ApprovementThread
}{
	Thread: structures.ApprovementThread{
		CoreMajorVersion: -1,
		Cache:            make(map[string]*structures.PoolStorage),
	},
}

var BLOCKS, EPOCH_DATA, APPROVEMENT_THREAD_METADATA, FINALIZATION_VOTING_STATS *leveldb.DB
