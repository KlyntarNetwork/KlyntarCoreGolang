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

	"github.com/KlyntarNetwork/KlyntarCoreGolang/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

type PivotSearchData struct {
	Position          int
	PivotPubKey       string
	FirstBlockByPivot *block.Block
	FirstBlockHash    string
}

type FirstBlockAssumption struct {
	IndexOfFirstBlockCreator int                                    `json:"indexOfFirstBlockCreator"`
	AfpForSecondBlock        structures.AggregatedFinalizationProof `json:"afpForSecondBlock"`
}

type FirstBlockResult struct {
	FirstBlockCreator, FirstBlockHash string
}

var CURRENT_PIVOT *PivotSearchData

func GetBlock(epochIndex uint, blockCreator string, index uint, epochHandler *structures.EpochHandler) *block.Block {

	blockID := fmt.Sprintf("%d:%s:%d", epochIndex, blockCreator, index)

	blockAsBytes, err := globals.BLOCKS.Get([]byte(blockID), nil)

	if err == nil {
		var blockParsed *block.Block
		err = json.Unmarshal(blockAsBytes, &blockParsed)
		if err == nil {
			return blockParsed
		}
	}

	// Find from other nodes

	quorumUrlsAndPubkeys := GetQuorumUrlsAndPubkeys(epochHandler)

	var quorumUrls []string

	for _, quorumMember := range quorumUrlsAndPubkeys {

		quorumUrls = append(quorumUrls, quorumMember.Url)

	}

	allKnownNodes := append(quorumUrls, globals.CONFIGURATION.BootstrapNodes...)

	type result struct {
		block *block.Block
	}

	resultChan := make(chan result, len(allKnownNodes))
	var wg sync.WaitGroup

	for _, node := range allKnownNodes {

		if node == globals.CONFIGURATION.MyHostname {
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

	localAfpAsBytes, err := globals.EPOCH_DATA.Get([]byte("AFP:"+blockID), nil)

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

func GetFirstBlockInEpoch(epochHandler *structures.EpochHandler) *FirstBlockResult {

	pivotData := CURRENT_PIVOT

	if pivotData == nil {

		allKnownNodes := GetQuorumUrlsAndPubkeys(epochHandler)

		var wg sync.WaitGroup

		responses := make(chan *FirstBlockAssumption, len(allKnownNodes))

		for _, node := range allKnownNodes {
			wg.Add(1)

			go func(nodeUrl string) {
				defer wg.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/first_block_assumption/%d", nodeUrl, epochHandler.Id), nil)
				if err != nil {
					return
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				var prop FirstBlockAssumption
				if err := json.NewDecoder(resp.Body).Decode(&prop); err != nil {
					return
				}

				responses <- &prop
			}(node.Url)
		}

		wg.Wait()
		close(responses)

		minimalIndexOfLeader := int(^uint(0) >> 1) // max int
		var afpForSecondBlock *structures.AggregatedFinalizationProof

		for prop := range responses {
			if prop == nil {
				continue
			}

			if prop.IndexOfFirstBlockCreator < 0 || prop.IndexOfFirstBlockCreator >= len(epochHandler.LeadersSequence) {
				continue
			}

			firstBlockCreator := epochHandler.LeadersSequence[prop.IndexOfFirstBlockCreator]

			if VerifyAggregatedFinalizationProof(&prop.AfpForSecondBlock, epochHandler) {

				expectedSecondBlockID := fmt.Sprintf("%d:%s:1", epochHandler.Id, firstBlockCreator)

				if expectedSecondBlockID == prop.AfpForSecondBlock.BlockID &&
					prop.IndexOfFirstBlockCreator < minimalIndexOfLeader {

					minimalIndexOfLeader = prop.IndexOfFirstBlockCreator
					afpForSecondBlock = &prop.AfpForSecondBlock
				}
			}
		}

		if afpForSecondBlock != nil {

			position := minimalIndexOfLeader
			pivotPubKey := epochHandler.LeadersSequence[position]

			firstBlockByPivot := GetBlock(epochHandler.Id, pivotPubKey, uint(0), epochHandler)
			firstBlockHash := afpForSecondBlock.PrevBlockHash

			if firstBlockByPivot != nil && firstBlockHash == firstBlockByPivot.GetHash() {

				pivotData = &PivotSearchData{

					Position:          position,
					PivotPubKey:       pivotPubKey,
					FirstBlockByPivot: firstBlockByPivot,
					FirstBlockHash:    firstBlockHash,
				}
			}
		}
	}

	if pivotData != nil {

		// In pivot we have first block created in epoch by some pool

		// Try to move closer to the beginning of the epochHandler.leadersSequence to find the real first block

		// Based on ALRP in pivot block - find the real first block

		blockToEnumerateAlrp := pivotData.FirstBlockByPivot

		if pivotData.Position == 0 {

			defer func() {
				pivotData = nil
			}()

			return &FirstBlockResult{

				FirstBlockCreator: pivotData.PivotPubKey,
				FirstBlockHash:    pivotData.FirstBlockHash,
			}
		}

		for position := pivotData.Position - 1; position >= 0; position-- {

			previousPool := epochHandler.LeadersSequence[position]

			raw, ok := blockToEnumerateAlrp.ExtraData["aggregatedLeadersRotationProofs"]

			if !ok {
				continue
			}

			jsonBytes, err := json.Marshal(raw)
			if err != nil {
				continue
			}

			var proofs map[string]structures.AggregatedLeaderRotationProof

			if err := json.Unmarshal(jsonBytes, &proofs); err != nil {
				continue
			}

			leaderRotationProof, ok := proofs[previousPool]
			if !ok {
				continue
			}

			if position == 0 {

				defer func() {
					pivotData = nil
				}()

				if leaderRotationProof.SkipIndex == -1 {
					return &FirstBlockResult{
						FirstBlockCreator: pivotData.PivotPubKey,
						FirstBlockHash:    pivotData.FirstBlockHash,
					}
				} else {
					return &FirstBlockResult{
						FirstBlockCreator: previousPool,
						FirstBlockHash:    leaderRotationProof.FirstBlockHash,
					}
				}

			} else if leaderRotationProof.SkipIndex != -1 {

				// Found new potential pivot

				firstBlockByNewPivot := GetBlock(epochHandler.Id, previousPool, 0, epochHandler)

				if firstBlockByNewPivot == nil {
					return nil
				}

				if firstBlockByNewPivot.GetHash() == leaderRotationProof.FirstBlockHash {
					pivotData = &PivotSearchData{
						Position:          position,
						PivotPubKey:       previousPool,
						FirstBlockByPivot: firstBlockByNewPivot,
						FirstBlockHash:    leaderRotationProof.FirstBlockHash,
					}

					break // break cycle to run the cycle later with new pivot
				} else {
					return nil
				}
			}
		}
	}

	return nil

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

		copy(arrayForIteration, epochHandler.LeadersSequence[:position])

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
