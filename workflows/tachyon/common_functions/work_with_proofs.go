package common_functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

type PivotSearchData struct {
	Position          int
	PivotPubKey       string
	FirstBlockByPivot *block.Block
	FirstBlockHash    string
}

var CURRENT_PIVOT *PivotSearchData

func GetBlock(epochIndex uint, blockCreator string, index uint, epochHandler *structures.EpochHandler) *block.Block {

	blockID := fmt.Sprintf("%d:%s:%d", epochIndex, blockCreator, index)

	blockAsBytes, err := tachyon.BLOCKS.Get([]byte(blockID), nil)

	if err == nil {
		var blockParsed *block.Block
		err = json.Unmarshal(blockAsBytes, &blockParsed)
		if err == nil {
			return blockParsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/block/%s", tachyon.CONFIGURATION.GetBlocksURL, blockID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err == nil {

		resp, err := http.DefaultClient.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {

			defer resp.Body.Close()

			var block block.Block

			if err := json.NewDecoder(resp.Body).Decode(&block); err == nil {

				return &block

			}

		}

	}

	// Find from other nodes

	quorumUrlsAndPubkeys := GetQuorumUrlsAndPubkeys(epochHandler)

	var quorumUrls []string

	for _, quorumMember := range quorumUrlsAndPubkeys {

		quorumUrls = append(quorumUrls, quorumMember.Url)

	}

	allKnownNodes := append(quorumUrls, tachyon.CONFIGURATION.BootstrapNodes...)

	type result struct {
		block *block.Block
	}

	resultChan := make(chan result, len(allKnownNodes))
	var wg sync.WaitGroup

	for _, node := range allKnownNodes {

		if node == tachyon.CONFIGURATION.MyHostname {
			continue
		}

		wg.Add(1)
		go func(endpoint string) {

			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			url := fmt.Sprintf("%s/block/%s", endpoint, blockID)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				return
			}
			defer resp.Body.Close()

			var block block.Block

			if err := json.NewDecoder(resp.Body).Decode(&block); err == nil {
				resultChan <- result{block: &block}
			}

		}(node)

	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.block != nil {
			return res.block
		}
	}

	return nil
}

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

	return uint(okSignatures) >= majority

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

func GetVerifiedAggregatedFinalizationProofByBlockId(blockID string, epochHandler *structures.EpochHandler) *structures.AggregatedFinalizationProof {

	localAfpAsBytes, err := tachyon.EPOCH_DATA.Get([]byte("AFP:"+blockID), nil)

	if err == nil {

		var localAfpParsed *structures.AggregatedFinalizationProof

		err = json.Unmarshal(localAfpAsBytes, &localAfpParsed)

		if err != nil {
			return nil
		}

		return localAfpParsed

	}

	quorum := GetQuorumUrlsAndPubkeys(epochHandler)

	resultChan := make(chan *structures.AggregatedFinalizationProof, len(quorum))

	var wg sync.WaitGroup

	for _, node := range quorum {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/aggregated_finalization_proof/"+blockID, nil)
			if err != nil {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			var afp structures.AggregatedFinalizationProof
			if err := json.NewDecoder(resp.Body).Decode(&afp); err != nil {
				return
			}

			if VerifyAggregatedFinalizationProof(&afp, epochHandler) {
				resultChan <- &afp
			}
		}(node.Url)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Return first valid AFP
	for res := range resultChan {
		if res != nil {
			return res
		}
	}

	return nil
}

func GetFirstBlockInEpoch(epochHandler *structures.EpochHandler) {

	pivotData := CURRENT_PIVOT

	if pivotData == nil {

	}

}

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

func CheckAlrpChainValidity(firstBlockInThisEpochByPool *block.Block, epochHandler *structures.EpochHandler, position int) bool {

	if aggregatedLeadersRotationProofsRef, ok := firstBlockInThisEpochByPool.ExtraData["aggregatedLeadersRotationProofs"]; ok {

		jsonBytes, err := json.Marshal(aggregatedLeadersRotationProofsRef)

		if err != nil {
			return false
		}

		var aggregatedLeadersRotationProofsParsed map[string]structures.AggregatedLeaderRotationProof

		if err := json.Unmarshal(jsonBytes, &aggregatedLeadersRotationProofsParsed); err != nil {
			return false
		}

		arrayIndexer := 0

		arrayForIteration := make([]string, position)

		copy(arrayForIteration, epochHandler.LeaderSequence[:position])

		// Reverse slice
		for i, j := 0, len(arrayForIteration)-1; i < j; i, j = i+1, j-1 {
			arrayForIteration[i], arrayForIteration[j] = arrayForIteration[j], arrayForIteration[i]
		}

		bumpedWithPoolWhoCreatedAtLeastOneBlock := false

		for _, poolPubKey := range arrayForIteration {

			if alrpForThisPool, ok := aggregatedLeadersRotationProofsParsed[poolPubKey]; ok {

				signaIsOk := VerifyAggregatedLeaderRotationProof(poolPubKey, &alrpForThisPool, epochHandler)

				if signaIsOk {

					arrayIndexer++

					if alrpForThisPool.SkipIndex >= 0 {

						bumpedWithPoolWhoCreatedAtLeastOneBlock = true

						break

					}

				} else {

					return false

				}

			} else {

				return false

			}

		}

		if arrayIndexer == position || bumpedWithPoolWhoCreatedAtLeastOneBlock {

			return true

		}

		return false

	}

	return false

}
