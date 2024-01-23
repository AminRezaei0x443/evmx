package main

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	ec "evmx/core"
	"evmx/state"
	"evmx/types"
	"evmx/vm"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
)

var (
	testHash    = common.BytesToHash([]byte("stuff"))
	fromAddress = common.BytesToAddress([]byte("a"))
	toAddress   = common.BytesToAddress([]byte("b"))
	amount      = big.NewInt(0)
	nonce       = uint64(0)
	gasLimit    = big.NewInt(100000)
	coinbase    = fromAddress
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func loadBin(filename string) []byte {
	code, err := ioutil.ReadFile(filename)
	must(err)
	return hexutil.MustDecode("0x" + string(code))
	//return []byte("0x" + string(code))
}
func loadAbi(filename string) abi.ABI {
	abiFile, err := os.Open(filename)
	must(err)
	defer abiFile.Close()
	abiObj, err := abi.JSON(abiFile)
	must(err)
	return abiObj
}

func main() {
	abiFileName := "./data/coin_sol_Coin.abi"
	binFileName := "./data/coin_sol_Coin.bin"
	data := loadBin(binFileName)

	msg := ec.Message{
		From:              fromAddress,
		To:                &toAddress,
		Nonce:             nonce,
		Data:              data,
		GasLimit:          gasLimit.Uint64(),
		GasPrice:          big.NewInt(0),
		SkipAccountChecks: true,
		GasFeeCap:         big.NewInt(0),
		GasTipCap:         big.NewInt(0),
	}

	cc := ChainContext{}
	block_ctx := ec.NewEVMBlockContext(cc.GetHeader(testHash, 0), cc, &fromAddress)
	tx_ctx := ec.NewEVMTxContext(&msg)

	dataPath := "/tmp/a.txt"
	os.Remove(dataPath)
	mdb, err := rawdb.NewLevelDBDatabase(dataPath, 100, 100, "idk", false)
	must(err)
	db := state.NewDatabase(mdb)

	root := common.Hash{}
	statedb, err := state.New(root, db, nil)
	must(err)
	//set balance
	//statedb.
	//statedb.GetOrNewStateObject(fromAddress)
	//statedb.GetOrNewStateObject(toAddress)
	statedb.AddBalance(fromAddress, big.NewInt(1e18))
	testBalance := statedb.GetBalance(fromAddress)
	fmt.Println("init testBalance =", testBalance)
	must(err)

	//	config := params.TestnetChainConfig
	config := params.MainnetChainConfig
	//cfg := logger.Config{}
	//structLogger := logger.NewStructLogger(&cfg)
	vmConfig := vm.Config{} //Tracer: structLogger
	/*, JumpTable: vm.NewByzantiumInstructionSet()*/

	evm := vm.NewEVM(block_ctx, tx_ctx, statedb, config, vmConfig)
	contractRef := vm.AccountRef(fromAddress)
	contractCode, contractAddr, gasLeftover, vmerr := evm.Create(contractRef, data, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	fmt.Printf("getcode:%x\n%x\n", contractCode, statedb.GetCode(contractAddr))

	statedb.SetBalance(fromAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = statedb.GetBalance(fromAddress)
	fmt.Println("after create contract, testBalance =", testBalance)
	abiObj := loadAbi(abiFileName)

	input, err := abiObj.Pack("minter")
	must(err)
	outputs, gasLeftover, vmerr := evm.Call(contractRef, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	//fmt.Printf("minter is %x\n", common.BytesToAddress(outputs))
	//fmt.Printf("call address %x\n", contractRef)

	sender := common.BytesToAddress(outputs)

	if !bytes.Equal(sender.Bytes(), fromAddress.Bytes()) {
		fmt.Println("caller are not equal to minter!!")
		os.Exit(-1)
	}

	senderAcc := vm.AccountRef(sender)

	input, err = abiObj.Pack("mint", sender, big.NewInt(1000000))
	must(err)
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	statedb.SetBalance(fromAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = evm.StateDB.GetBalance(fromAddress)

	input, err = abiObj.Pack("send", toAddress, big.NewInt(11))
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	//send
	input, err = abiObj.Pack("send", toAddress, big.NewInt(19))
	must(err)
	outputs, gasLeftover, vmerr = evm.Call(senderAcc, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	fmt.Printf("toAddress %x\n", toAddress)
	// get balance
	input, err = abiObj.Pack("balances", toAddress)
	must(err)
	outputs, gasLeftover, vmerr = evm.Call(contractRef, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)
	Print(outputs, "balances")

	// get balance
	input, err = abiObj.Pack("balances", sender)
	must(err)
	outputs, gasLeftover, vmerr = evm.Call(contractRef, contractAddr, input, statedb.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)
	Print(outputs, "balances")

	// get event
	logs := statedb.Logs()

	for _, log := range logs {
		fmt.Printf("%#v\n", log)
		for _, topic := range log.Topics {
			fmt.Printf("topic: %#v\n", topic)
		}
		fmt.Printf("data: %#v\n", log.Data)
	}

	root, err = statedb.Commit(0, true)
	must(err)
	err = db.TrieDB().Commit(root, true)
	must(err)

	fmt.Println("Root Hash", root.Hex())
	mdb.Close()
	//
	//mdb2, err := ethdb.NewLDBDatabase(dataPath, 100, 100)
	//defer mdb2.Close()
	//must(err)
	//db2 := state.NewDatabase(mdb2)
	//statedb2, err := state.New(root, db2)
	//must(err)
	//testBalance = statedb2.GetBalance(fromAddress)
	//fmt.Println("get testBalance =", testBalance)
	//if !bytes.Equal(contractCode, statedb2.GetCode(contractAddr)) {
	//	fmt.Println("BUG!,the code was changed!")
	//	os.Exit(-1)
	//}
	//getVariables(statedb2, contractAddr)
}

//
//func getVariables(statedb *state.StateDB, hash common.Address) {
//	cb := func(key, value common.Hash) bool {
//		fmt.Printf("key=%x,value=%x\n", key, value)
//		return true
//	}
//
//	statedb.ForEachStorage(hash, cb)
//
//}

func Print(outputs []byte, name string) {
	fmt.Printf("method=%s, output=%x\n", name, outputs)
}

type ChainContext struct{}

func (cc ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {

	return &types.Header{
		// ParentHash: common.Hash{},
		// UncleHash:  common.Hash{},
		Coinbase: fromAddress,
		//	Root:        common.Hash{},
		//	TxHash:      common.Hash{},
		//	ReceiptHash: common.Hash{},
		//	Bloom:      types.BytesToBloom([]byte("duanbing")),
		Difficulty: big.NewInt(1),
		Number:     big.NewInt(1),
		GasLimit:   1000000,
		GasUsed:    0,
		Time:       uint64(time.Now().Unix()),
		Extra:      nil,
		//MixDigest:  testHash,
		//Nonce:      types.EncodeNonce(1),
	}
}
