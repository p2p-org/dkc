package utils

import (
	"context"
	"io/ioutil"

	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type DirkStore struct {
	Location string
	Wallets  []types.Wallet
}

type Peers = map[uint64]string

type Account struct {
	ID        uint64
	Key       []byte
	Signature []byte
}

func CreateStore(path string) (types.Store, error) {
	store := filesystem.New(filesystem.WithLocation(path))
	return store, nil
}

func LoadStores(ctx context.Context, walletDir string, passphrases [][]byte) ([]DirkStore, error) {
	var stores []DirkStore

	dirs, err := ioutil.ReadDir(walletDir)
	if err != nil {
		return nil, errors.Wrap(err, ErrorWalletDirWrapper)
	}
	for _, f := range dirs {
		if f.IsDir() {
			store, err := LoadStore(ctx, walletDir+"/"+f.Name(), passphrases)
			if err != nil {
				return nil, errors.Wrap(err, ErrorLoadStoreWrapper)
			}
			stores = append(stores, *store)
		}
	}
	return stores, nil
}

func LoadStore(ctx context.Context, location string, passphrases [][]byte) (*DirkStore, error) {
	dirkStore := DirkStore{}
	dirkStore.Location = location
	var wallets []types.Wallet
	store := filesystem.New(filesystem.WithLocation(location))
	if err := e2wallet.UseStore(store); err != nil {
		return nil, errors.Wrap(err, ErrorUseStoreWrapper)
	}
	for wallet := range e2wallet.Wallets() {
		wallets = append(wallets, wallet)
	}
	dirkStore.Wallets = wallets
	return &dirkStore, nil
}
