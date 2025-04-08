package globals

//_____________________________________________________DEFINE GLOBAL ACCESS VALUES____________________________________________________

// Pathes to 3 main direcories
var CHAINDATA_PATH, GENESIS_PATH, CONFIGS_PATH string

// Global configs (resolved by <CONFIGS_PATH>, example available in workflows/tachyon/templates/configs.json)
var CONFIGS map[string]interface{}

// Load genesis from JSON file to pre-set the state
var GENESIS map[string]interface{}
