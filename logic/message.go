package logic

import (
	ec "github.com/AminRezaei0x443/evmx/core"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const HardCodedGasLimit uint64 = 10000

func NewMessage(from common.Address, to common.Address, nonce uint64, value uint64, data []byte) *ec.Message {
	return &ec.Message{
		From:              from,
		To:                &to,
		Nonce:             nonce,
		Data:              data,
		Value:             big.NewInt(0).SetUint64(value),
		GasLimit:          HardCodedGasLimit,
		GasPrice:          big.NewInt(0),
		SkipAccountChecks: true,
		GasFeeCap:         big.NewInt(0),
		GasTipCap:         big.NewInt(0),
	}
}
