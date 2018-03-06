package main

import (
	"fmt"
	"time"
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func waitDeployed(backend bind.DeployBackend, ctx context.Context, tx *types.Transaction) common.Address {
	mined := make(chan common.Address)
	go func() {
		fmt.Printf("Wait deployed... <= %s\n", tx.Hash().Hex())
		address, _ := bind.WaitDeployed(ctx, backend, tx)
		mined <- address
		close(mined)
	}()
	select {
	case a := <-mined:
		fmt.Printf("Contract %s deployed.\n", a.String())
		return a
	case <-time.After(20 * time.Second):
		panic(fmt.Errorf("%s timeout", tx.Hash().String()))
	}
}

func waitTx(backend bind.DeployBackend, ctx context.Context, tx *types.Transaction) types.Receipt {
	mined := make(chan types.Receipt)
	go func() {
		fmt.Printf("Wait tx %s ...", tx.Hash().Hex())
		receipt, err := bind.WaitMined(ctx, backend, tx)
		if err != nil {
			panic(err)
		}
		mined <- *receipt
		close(mined)
	}()
	select {
	case r := <-mined:
		fmt.Printf("Tx Mined! GasUsed: %d\n", r.GasUsed)
		return r
	case <-time.After(20 * time.Second):
		panic(fmt.Errorf("%s timeout", tx.Hash().String()))
	}
}
