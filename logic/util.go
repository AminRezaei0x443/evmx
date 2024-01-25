package logic

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"os"
)

func LoadContractCode(fn string) ([]byte, error) {
	code, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return hexutil.MustDecode("0x" + string(code)), nil
}

func LoadContractABI(fn string) (*abi.ABI, error) {
	abiFile, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer abiFile.Close()
	abiObj, err := abi.JSON(abiFile)
	if err != nil {
		return nil, err
	}
	return &abiObj, nil
}

type LoadedContract struct {
	code    []byte
	abi     *abi.ABI
	address *common.Address
}

func LoadContract(abiFn string, binFn string) (lc *LoadedContract, e error) {
	c := LoadedContract{}
	cc, e := LoadContractCode(binFn)
	if e != nil {
		return
	}
	ca, e := LoadContractABI(abiFn)
	if e != nil {
		return
	}
	c.code = cc
	c.abi = ca
	c.address = nil
	return &c, nil
}

func LoadContract2(abiFn string, binFn string, addr common.Address) (lc *LoadedContract, e error) {
	c, e := LoadContract(abiFn, binFn)
	if e != nil {
		return
	}
	c.address = &addr
	return c, nil
}

func (lc *LoadedContract) Code() []byte {
	return lc.code
}

func (lc *LoadedContract) Abi() *abi.ABI {
	return lc.abi
}

func (lc *LoadedContract) Addr() common.Address {
	return *lc.address
}
