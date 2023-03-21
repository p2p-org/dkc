package combine

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type DirkStore struct {
	Location string
	Wallets  []e2wtypes.Wallet
}

type Peers = map[uint64]string

type Accounts = map[string][][]byte

func loadStore(ctx context.Context, location string) (*DirkStore, error) {
	dirkStore := DirkStore{}
	dirkStore.Location = location
	var wallets []e2wtypes.Wallet
	store := filesystem.New(filesystem.WithLocation(location))
	if err := e2wallet.UseStore(store); err != nil {
		return nil, errors.Wrap(err, "failed to use store")
	}
	for wallet := range e2wallet.Wallets() {
		wallets = append(wallets, wallet)
	}
	dirkStore.Wallets = wallets
	return &dirkStore, nil
}

func loadStores(ctx context.Context) ([]DirkStore, error) {
	var stores []DirkStore
	walletDir := viper.GetString("walletDir")
	dirs, err := ioutil.ReadDir(walletDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wallet dir")
	}
	for _, f := range dirs {
		if f.IsDir() {
			store, err := loadStore(ctx, walletDir+"/"+f.Name())
			if err != nil {
				return nil, errors.Wrap(err, "can't load store")
			}
			stores = append(stores, *store)
		}
	}
	return stores, nil
}

func combineWallets(ctx context.Context) (Accounts, error) {
	stores, err := loadStores(ctx)
	if err != nil {
		fmt.Println(err)
	}

	var peers Peers
	participantsIDs := make([]uint64, 0)
	err = viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	accountDatas := make(Accounts)

	for _, store := range stores {
		for id := range peers {
			peerExists, _ := regexp.MatchString(filepath.Base(store.Location)+":.*", peers[id])
			if peerExists {
				participantsIDs = append(participantsIDs, id)
			}
		}
		for _, wallet := range store.Wallets {
			for account := range wallet.Accounts(ctx) {
				key, err := getAccountKey(ctx, account)
				if err != nil {
					fmt.Println("Error")
				}

				accountDatas[account.Name()] = append(
					accountDatas[account.Name()],
					key,
				)
			}
		}
	}

	for _, account := range accountDatas {
		_, _ = bls.Recover(ctx, account, participantsIDs)
	}

	return accountDatas, nil
}
