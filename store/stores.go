package store

import (
	"context"

	"github.com/pkg/errors"
)

type InputStore interface {
	// Get Wallets Names And Account Names
	GetWalletsAccountsMap() ([]AccountsData, []string, error)
	// Get Private Key From Wallet Using Account Name
	GetPrivateKey(walletName string, accountName string) ([]byte, error)
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
}

type OutputStore interface {
	// Create Store
	Create() error
	// Create New Wallet
	CreateWallet(name string) error
	// Save Private Key To Wallet
	SavePrivateKey(walletName string, accountName string, privateKey []byte) error
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
}

func InputStoreInit(ctx context.Context, storeType string) (InputStore, error) {
	switch storeType {
	case "distributed":
		store, err := newDistributedStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as input store")
		}
		store.Ctx = ctx
		return &store, nil
	case "hierarchical deterministic":
		store, err := newHDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init hierarchial deterministic store as input store")
		}
		store.Ctx = ctx
		return &store, nil
	case "non-deterministic":
		store, err := newNDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as input store")
		}
		store.Ctx = ctx
		return &store, nil
	case "keystore":
		store, err := newKeystoreStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init keystore store as input store")
		}
		store.Ctx = ctx
		return &store, nil
	default:
		return nil, errors.New("incorrect input wallet type")
	}

}

func OutputStoreInit(ctx context.Context, storeType string) (OutputStore, error) {
	switch storeType {
	case "distributed":
		store, err := newDistributedStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	case "non-deterministic":
		store, err := newNDStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	case "keystore":
		store, err := newKeystoreStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init keystore store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	default:
		return nil, errors.New("incorrect output wallet type")
	}

}
