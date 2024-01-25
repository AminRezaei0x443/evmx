package logic

import (
	ec "evmx/core"
	"evmx/rawdb"
	"evmx/state"
	"evmx/vm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"os"
)

type EVMSession struct {
	dataPath   string
	cleanStart bool
	db         state.Database
	State      *state.StateDB
	VM         *vm.EVM
}

type DeployResult struct {
	Ok                 bool
	Error              error
	ContractAddr       common.Address
	ContractCode       []byte
	DeployerNewBalance uint64
}

type CallResult struct {
	Ok               bool
	Error            error
	Output           []byte
	CallerNewBalance uint64
}

func NewEVMSession(dataPath string, cleanStart bool) *EVMSession {
	es := &EVMSession{
		dataPath:   dataPath,
		cleanStart: cleanStart,
	}
	return es
}

func (es *EVMSession) OpenDB() (e error) {
	if es.cleanStart {
		e = os.Remove(es.dataPath)
		if e != nil {
			return
		}
	}
	mdb, e := rawdb.NewLevelDBDatabase(es.dataPath, 100, 100, "idk", false)
	if e != nil {
		return
	}
	es.db = state.NewDatabase(mdb)
	return nil
}

func (es *EVMSession) OpenState(root common.Hash) (e error) {
	es.State, e = state.New(root, es.db, nil)
	return
}

func (es *EVMSession) EmptyState() error {
	return es.OpenState(common.Hash{})
}

func (es *EVMSession) InitEVM(msg *ec.Message) error {
	cc := MockedChainContext{}
	dummyHash := common.BytesToHash([]byte("TEST"))

	blockCtx := ec.NewEVMBlockContext(cc.GetHeader(dummyHash, 0), cc, &msg.From)
	txCtx := ec.NewEVMTxContext(msg)

	config := params.MainnetChainConfig
	vmConfig := vm.Config{}

	es.VM = vm.NewEVM(blockCtx, txCtx, es.State, config, vmConfig)

	return nil
}

func (es *EVMSession) DeployContract(from common.Address, contract *LoadedContract) *DeployResult {
	result := DeployResult{}
	deployer := vm.AccountRef(from)
	deployerBalance := es.BalanceOf(from)
	code, addr, gasLeftover, e := es.VM.Create(deployer, contract.code, deployerBalance, big.NewInt(0))
	if e != nil {
		result.Ok = false
		result.Error = e
		return &result
	}
	es.SetBalance(from, gasLeftover)
	result.Ok = true
	result.Error = nil
	result.ContractCode = code
	result.ContractAddr = addr
	result.DeployerNewBalance = gasLeftover
	contract.address = &addr
	return &result
}

func (es *EVMSession) BalanceOf(target common.Address) uint64 {
	balance := es.State.GetBalance(target).Uint64()
	return balance
}

func (es *EVMSession) SetBalance(target common.Address, value uint64) {
	es.State.SetBalance(target, big.NewInt(0).SetUint64(value))
}

func (es *EVMSession) CallContract(caller common.Address, contract common.Address, input []byte) *CallResult {
	result := CallResult{}

	balance := es.BalanceOf(caller)
	ref := vm.AccountRef(caller)
	outputs, gasLeftover, e := es.VM.Call(ref, contract, input, balance, big.NewInt(0))
	if e != nil {
		result.Ok = false
		result.Error = e
		return &result
	}

	es.SetBalance(caller, gasLeftover)
	result.Ok = true
	result.Error = nil
	result.Output = outputs
	result.CallerNewBalance = gasLeftover
	return &result
}

func (es *EVMSession) CallContractArgs(caller common.Address, contract *LoadedContract, name string, args ...interface{}) *CallResult {
	result := CallResult{}

	input, e := contract.abi.Pack(name, args...)
	if e != nil {
		result.Ok = false
		result.Error = e
		return &result
	}

	return es.CallContract(caller, *contract.address, input)
}

func (es *EVMSession) CommitState(block uint64) (root common.Hash, e error) {
	root, e = es.State.Commit(block, true)
	if e != nil {
		return
	}
	e = es.db.TrieDB().Commit(root, true)
	return
}
