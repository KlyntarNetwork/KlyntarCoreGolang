package globals

import (
	"os"
	"strconv"
	"sync"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/transaction"
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

func openDB(dbName string) *leveldb.DB {
	db, err := leveldb.OpenFile(CHAINDATA_PATH+"/"+dbName, nil)
	if err != nil {
		panic("Impossible to open db : " + dbName + " =>" + err.Error())
	}
	return db
}

var CONFIGURATION structures.NodeLevelConfig
var GENESIS structures.Genesis

var MEMPOOL struct {
	slice []transaction.Transaction
	mutex sync.RWMutex
}

var APPROVEMENT_THREAD_CACHE = make(map[string]*structures.Pool)

var FINALIZATION_PROOFS_CACHE = make(map[string]map[string]string)

var TEMP_CACHE = make(map[string]any)

var GENERATION_THREAD structures.GenerationThread

var APPROVEMENT_THREAD structures.ApprovementThread

var (
	BLOCKS                      = openDB("BLOCKS")
	EPOCH_DATA                  = openDB("EPOCH_DATA")
	APPROVEMENT_THREAD_METADATA = openDB("APPROVEMENT_THREAD_METADATA")
	FINALIZATION_VOTING_STATS   = openDB("FINALIZATION_VOTING_STATS")
)

var VOTING_REQUESTS chan struct{}
