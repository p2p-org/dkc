package utils

import (
	"github.com/google/uuid"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type NDWallet interface {
	types.WalletAccountImporter
	types.WalletLocker
}

type DWallet interface {
	types.WalletDistributedAccountImporter
	types.WalletLocker
}

func CreateNDWallet(store types.Store) (ndWallet NDWallet) {
	wallet, err := createWallet(store, "non-deterministic")
	if err != nil {
		panic(err)
	}
	ndWallet = wallet.(NDWallet)

	return
}

func CreateDWallet(store types.Store) (dWallet DWallet) {
	wallet, err := createWallet(store, "distributed")
	if err != nil {
		panic(err)
	}
	dWallet = wallet.(DWallet)

	return
}

func createWallet(store types.Store, wType string) (wallet types.Wallet, err error) {
	e2wallet.UseStore(store)
	wallet, err = e2wallet.CreateWallet(uuid.New().String(), e2wallet.WithType(wType))
	if err != nil {
		return
	}
	return
}
