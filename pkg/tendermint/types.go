package tendermint

import "time"

type StatusResponse struct {
	Result StatusResult `json:"result"`
}

type StatusResult struct {
	NodeInfo      NodeInfo      `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}

type NodeInfo struct {
	Moniker string `json:"moniker"`
}

type SyncInfo struct {
	LatestBlockTime time.Time `json:"latest_block_time"`
	CatchingUp      bool      `json:"catching_up"`
}

type ValidatorInfo struct {
	VotingPower string `json:"voting_power"`
}

type BlockResponse struct {
	Result BlockResult `json:"result"`
}

type BlockResult struct {
	Block Block `json:"block"`
}

type Block struct {
	Header BlockHeader `json:"header"`
}

type BlockHeader struct {
	Height string    `json:"height"`
	Time   time.Time `json:"time"`
}
