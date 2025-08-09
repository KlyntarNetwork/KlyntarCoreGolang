package common_functions

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

func GetBlock(epochIndex int, blockCreator string, index uint, epochHandler *structures.EpochDataHandler) *block.Block {

	blockID := strconv.Itoa(epochIndex) + ":" + blockCreator + ":" + strconv.Itoa(int(index))

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

	resultChan := make(chan *block.Block, len(allKnownNodes))
	var wg sync.WaitGroup

	for _, node := range allKnownNodes {

		if node == globals.CONFIGURATION.MyHostname {
			continue
		}

		wg.Add(1)
		go func(endpoint string) {

			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			url := endpoint + "/block/" + blockID
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
				resultChan <- &block
			}

		}(node)

	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for block := range resultChan {
		if block != nil {
			return block
		}
	}

	return nil
}

func VerifyAggregatedEpochFinalizationProof(
	proofStruct *structures.AggregatedEpochFinalizationProof,
	quorum []string,
	majority int,
	epochFullID string,
) bool {

	dataThatShouldBeSigned := "EPOCH_DONE:" +
		strconv.Itoa(int(proofStruct.LastLeader)) + ":" +
		strconv.Itoa(int(proofStruct.LastIndex)) + ":" +
		proofStruct.LastHash + ":" +
		proofStruct.HashOfFirstBlockByLastLeader + ":" +
		epochFullID

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

	return okSignatures >= majority

}

func VerifyAggregatedFinalizationProof(
	proof *structures.AggregatedFinalizationProof,
	epochHandler *structures.EpochDataHandler,
) bool {

	epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochHandler.Id)

	dataThatShouldBeSigned := proof.PrevBlockHash + proof.BlockId + proof.BlockHash + epochFullID

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

	return okSignatures >= majority
}

func VerifyAggregatedLeaderRotationProof(
	pubKeyOfSomePreviousLeader string,
	proof *structures.AggregatedLeaderRotationProof,
	epochHandler *structures.EpochDataHandler,
) bool {

	epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochHandler.Id)

	dataThatShouldBeSigned := "LEADER_ROTATION_PROOF:" + pubKeyOfSomePreviousLeader + ":" +
		proof.FirstBlockHash + ":" +
		strconv.Itoa(proof.SkipIndex) + ":" +
		proof.SkipHash + ":" +
		epochFullID

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

	return okSignatures >= majority

}

func CheckAlrpChainValidity(firstBlockInThisEpochByPool *block.Block, epochHandler *structures.EpochDataHandler, position int) bool {

	aggregatedLeadersRotationProofsRef := firstBlockInThisEpochByPool.ExtraData.AggregatedLeadersRotationProofs

	arrayIndexer := 0

	arrayForIteration := slices.Clone(epochHandler.LeadersSequence[:position])

	slices.Reverse(arrayForIteration) // we need reversed version

	bumpedWithPoolWhoCreatedAtLeastOneBlock := false

	for _, poolPubKey := range arrayForIteration {

		if alrpForThisPool, ok := aggregatedLeadersRotationProofsRef[poolPubKey]; ok {

			signaIsOk := VerifyAggregatedLeaderRotationProof(poolPubKey, alrpForThisPool, epochHandler)

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

func GetVerifiedAggregatedFinalizationProofByBlockId(blockID string, epochHandler *structures.EpochDataHandler) *structures.AggregatedFinalizationProof {

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

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
