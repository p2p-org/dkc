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
	walletName := uuid.New().String()
	wallet, err := createWallet(store, "non-deterministic", walletName)
	if err != nil {
		return nil, err
	}
	ndWallet := wallet.(NDWallet)

	return ndWallet, err
}

func CreateDWallet(store types.Store, walletName string) (DWallet, error) {
	wallet, err := createWallet(store, "distributed", walletName)
	if err != nil {
		return nil, err
	}
	dWallet := wallet.(DWallet)

	return dWallet, nil
}

func createWallet(store types.Store, wType string, walletName string) (types.Wallet, error) {
	err := e2wallet.UseStore(store)
	if err != nil {
		return nil, err
	}
	wallet, err := e2wallet.CreateWallet(walletName, e2wallet.WithType(wType))
	if err != nil {
		return nil, errors.Wrap(err, errorFailedToCreateWalletWrapper)
	}
	return wallet, nil
}
