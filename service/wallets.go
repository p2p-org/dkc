package service

import (
	"github.com/google/uuid"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func CreateWallet(store types.Store, wType string) (wallet types.Wallet) {
	e2wallet.UseStore(store)
	wallet, err := e2wallet.CreateWallet(uuid.New().String(), e2wallet.WithType(wType))
	if err != nil {
		panic(err)
	}
	return
}
