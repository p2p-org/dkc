package utils

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

const (
	errorFailedToCreateWalletWrapper = "failed to create wallet"
)

type NDWallet interface {
	types.WalletAccountImporter
	types.WalletLocker
}

type DWallet interface {
	types.WalletDistributedAccountImporter
	types.WalletLocker
}

func CreateNDWallet(store types.Store) (NDWallet, error) {
	wallet, err := createWallet(store, "non-deterministic")
	if err != nil {
		return nil, err
	}
	ndWallet := wallet.(NDWallet)

	return ndWallet, err
}

func CreateDWallet(store types.Store) (DWallet, error) {
	wallet, err := createWallet(store, "distributed")
	if err != nil {
		return nil, err
	}
	dWallet := wallet.(DWallet)

	return dWallet, nil
}

func createWallet(store types.Store, wType string) (types.Wallet, error) {
	e2wallet.UseStore(store)
	wallet, err := e2wallet.CreateWallet(uuid.New().String(), e2wallet.WithType(wType))
	if err != nil {
		return nil, errors.Wrap(err, errorFailedToCreateWalletWrapper)
	}
	return wallet, nil
}
