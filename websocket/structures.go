package websocket

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
)

type WsLeaderRotationProofRequest struct {
	Route               string                                 `json:"route"`
	AfpForFirstBlock    structures.AggregatedFinalizationProof `json:"afpForFirstBlock"`
	IndexOfPoolToRotate int                                    `json:"hisIndexInLeadersSequence"`
	SkipData            structures.PoolVotingStat              `json:"skipData"`
}

type WsLeaderRotationProofResponseOk struct {
	Route         string `json:"route"`
	Voter         string `json:"voter"`
	ForPoolPubkey string `json:"forPoolPubkey"`
	Type          string `json:"type"`
	Sig           string `json:"sig"`
}

type WsLeaderRotationProofResponseUpgrade struct {
	Route            string                                 `json:"route"`
	Voter            string                                 `json:"voter"`
	ForPoolPubkey    string                                 `json:"forPoolPubkey"`
	Type             string                                 `json:"type"`
	AfpForFirstBlock structures.AggregatedFinalizationProof `json:"afpForFirstBlock"`
	SkipData         structures.PoolVotingStat              `json:"skipData"`
}

type WsFinalizationProofRequest struct {
	Block            block.Block                            `json:"block"`
	PreviousBlockAfp structures.AggregatedFinalizationProof `json:"previousBlockAfp"`
}

type WsFinalizationProofResponse struct {
	Voter             string `json:"voter"`
	FinalizationProof string `json:"finalizationProof"`
	VotedForHash      string `json:"votedForHash"`
}
