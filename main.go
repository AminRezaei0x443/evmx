package main

import (
	"fmt"
	"github.com/AminRezaei0x443/evmx/logic"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func fatal(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	abiFn := "./data/coin_sol_Coin.abi"
	binFn := "./data/coin_sol_Coin.bin"

	loaded, e := logic.LoadContract(abiFn, binFn)
	fatal(e)

	from := common.BytesToAddress([]byte("alice"))
	to := common.BytesToAddress([]byte("bob"))

	fmt.Printf("alice: %x\n", from)
	fmt.Printf("bob: %x\n", to)

	msg := logic.NewMessage(from, to, 0, 0, loaded.Code())

	session := logic.NewEVMSession("./tmp/x.txt", false)

	e = session.OpenDB()
	fatal(e)

	root := common.Hash{}
	e = session.OpenState(root)
	fatal(e)

	e = session.InitEVM(msg)
	fatal(e)

	session.SetBalance(from, 1e18)
	fmt.Println("balance(alice) =", session.BalanceOf(from))

	depResult := session.DeployContract(from, loaded)

	if !depResult.Ok {
		fatal(depResult.Error)
	}

	fmt.Println("Contract deployed at:", depResult.ContractAddr)
	fmt.Println("Deployer Balance:", depResult.DeployerNewBalance)

	// get minter value from contract
	callResult := session.CallContractArgs(from, loaded, "minter")

	if !callResult.Ok {
		fatal(callResult.Error)
	}

	fmt.Printf("minter: %x\n", common.BytesToAddress(callResult.Output))
	fmt.Println("New Balance:", callResult.CallerNewBalance)

	//senderAcc := vm.AccountRef(sender)

	// let's mint some stuff
	callResult = session.CallContractArgs(from, loaded, "mint", from, big.NewInt(1000000))

	if !callResult.Ok {
		fatal(callResult.Error)
	}

	fmt.Println("Called mint")
	fmt.Println("New Balance:", callResult.CallerNewBalance)

	// let's transfer some to bob
	callResult = session.CallContractArgs(from, loaded, "send", to, big.NewInt(11))

	if !callResult.Ok {
		fatal(callResult.Error)
	}

	fmt.Println("Called send")
	fmt.Println("New Balance:", callResult.CallerNewBalance)

	// get balances
	callResult = session.CallContractArgs(from, loaded, "balances", from)

	if !callResult.Ok {
		fatal(callResult.Error)
	}

	fmt.Println("Alice Coins:", callResult.Output)

	callResult = session.CallContractArgs(from, loaded, "balances", to)

	if !callResult.Ok {
		fatal(callResult.Error)
	}

	fmt.Println("Bob Coins:", callResult.Output)
	//
	//// get event
	//logs := statedb.Logs()
	//
	//for _, log := range logs {
	//	fmt.Printf("%#v\n", log)
	//	for _, topic := range log.Topics {
	//		fmt.Printf("topic: %#v\n", topic)
	//	}
	//	fmt.Printf("data: %#v\n", log.Data)
	//}

	root, e = session.CommitState(0)
	fmt.Println("Root Hash", root.Hex())
	//mdb.Close()
}
