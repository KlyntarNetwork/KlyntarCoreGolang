package common_functions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

func VerifyAggregatedEpochFinalizationProof(
	proofStruct *structures.AggregatedEpochFinalizationProof,
	quorum []string,
	majority uint,
	epochFullID string,
) bool {

	dataThatShouldBeSigned := fmt.Sprintf(
		"EPOCH_DONE:%d:%d:%s:%s:%s",
		proofStruct.LastLeader,
		proofStruct.LastIndex,
		proofStruct.LastHash,
		proofStruct.HashOfFirstBlockByLastLeader,
		epochFullID,
	)

	okSignatures := 0
	seen := make(map[string]bool)
	quorumMap := make(map[string]bool)

	for _, pk := range quorum {
		quorumMap[strings.ToLower(pk)] = true
	}

	for pubKey, signature := range proofStruct.Proofs {

		if ed25519.VerifySignature(dataThatShouldBeSigned, pubKey, signature) {
			loweredPubKey := strings.ToLower(pubKey)
			if quorumMap[loweredPubKey] && !seen[loweredPubKey] {
				seen[loweredPubKey] = true
				okSignatures++
			}
		}
	}

	if uint(okSignatures) >= majority {
		return true
	}

	return false
}

func VerifyAggregatedFinalizationProof(
	proof *structures.AggregatedFinalizationProof,
	epochHandler *structures.EpochHandler,
) bool {

	epochFullID := epochHandler.Hash + "#" + strconv.FormatUint(uint64(epochHandler.Id), 10)
	dataThatShouldBeSigned := proof.PrevBlockHash + proof.BlockID + proof.BlockHash + epochFullID

	majority := GetQuorumMajority(epochHandler)

	okSignatures := 0
	seen := make(map[string]bool)
	quorumMap := make(map[string]bool)

	for _, pk := range epochHandler.Quorum {
		quorumMap[strings.ToLower(pk)] = true
	}

	for pubKey, signature := range proof.Proofs {

		if ed25519.VerifySignature(dataThatShouldBeSigned, pubKey, signature) {
			loweredPubKey := strings.ToLower(pubKey)
			if quorumMap[loweredPubKey] && !seen[loweredPubKey] {
				seen[loweredPubKey] = true
				okSignatures++
			}
		}
	}

	return uint(okSignatures) >= majority
}

func GetVerifiedAggregatedFinalizationProofByBlockId(blockID string, epochHandler *structures.EpochHandler) {
}

func GetFirstBlockInEpoch(epochHandler *structures.EpochHandler) {}

func VerifyAggregatedLeaderRotationProof(
	pubKeyOfSomePreviousLeader string,
	proof *structures.AggregatedLeaderRotationProof,
	epochHandler *structures.EpochHandler,
) bool {

	epochFullID := epochHandler.Hash + "#" + strconv.FormatUint(uint64(epochHandler.Id), 10)

	dataThatShouldBeSigned := fmt.Sprintf(
		"LEADER_ROTATION_PROOF:%s:%s:%d:%s:%s",
		pubKeyOfSomePreviousLeader,
		proof.FirstBlockHash,
		proof.SkipIndex,
		proof.SkipHash,
		epochFullID,
	)

	majority := GetQuorumMajority(epochHandler)

	okSignatures := 0
	seen := make(map[string]bool)
	quorumMap := make(map[string]bool)

	for _, pk := range epochHandler.Quorum {
		quorumMap[strings.ToLower(pk)] = true
	}

	for pubKey, signature := range proof.Proofs {

		if ed25519.VerifySignature(dataThatShouldBeSigned, pubKey, signature) {
			loweredPubKey := strings.ToLower(pubKey)
			if quorumMap[loweredPubKey] && !seen[loweredPubKey] {
				seen[loweredPubKey] = true
				okSignatures++
			}
		}
	}

	return uint(okSignatures) >= majority

}

func CheckAlrpChainValidity() {}
