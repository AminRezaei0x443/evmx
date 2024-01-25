package logic

import (
	"evmx/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

type MockedChainContext struct{}

func (cc MockedChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	// TODO: This is mocked chain context, so it just returns simplest acceptable block header
	miner := common.BytesToAddress([]byte("random-miner"))
	return &types.Header{
		Coinbase:   miner,
		Difficulty: big.NewInt(1),
		Number:     big.NewInt(1),
		GasLimit:   1000000,
		GasUsed:    0,
		Time:       uint64(time.Now().Unix()),
		Extra:      nil,
	}
}
