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
	Network string `json:"network"`
	Version string `json:"version"`
}

type SyncInfo struct {
	LatestBlockHeight int64     `json:"latest_block_height,string"`
	LatestBlockTime   time.Time `json:"latest_block_time"`
	CatchingUp        bool      `json:"catching_up"`
}

type ValidatorInfo struct {
	VotingPower int64 `json:"voting_power,string"`
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
	Height int64     `json:"height,string"`
	Time   time.Time `json:"time"`
}

type AbciQueryResponse struct {
	Result AbciQueryResult `json:"result"`
}

type AbciQueryResult struct {
	Response AbciResponse `json:"response"`
}

type AbciResponse struct {
	Code  int    `json:"code"`
	Value []byte `json:"value"`
}

type BlocksInfo struct {
	NewerBlock BlockResponse
	OlderBlock BlockResponse
}

func (b *BlocksInfo) BlockTime() float64 {
	blocksDiffHeight := b.NewerBlock.Result.Block.Header.Height - b.OlderBlock.Result.Block.Header.Height
	blocksDiffTime := b.NewerBlock.Result.Block.Header.Time.Sub(b.OlderBlock.Result.Block.Header.Time)
	return blocksDiffTime.Seconds() / float64(blocksDiffHeight)
}
