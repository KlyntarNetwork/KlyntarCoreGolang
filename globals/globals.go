package globals

import (
	"os"
	"strconv"
	"sync"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/syndtr/goleveldb/leveldb"
)

func GetCoreMajorVersion(versionFilePath string) (int, error) {

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
	Slice []structures.Transaction
	Mutex sync.Mutex
}

var GENERATION_THREAD structures.GenerationThread

var APPROVEMENT_THREAD struct {
	RWMutex sync.RWMutex
	Thread  structures.ApprovementThread
}

var (
	BLOCKS                      = openDB("BLOCKS")
	EPOCH_DATA                  = openDB("EPOCH_DATA")
	APPROVEMENT_THREAD_METADATA = openDB("APPROVEMENT_THREAD_METADATA")
	FINALIZATION_VOTING_STATS   = openDB("FINALIZATION_VOTING_STATS")
)
