package combine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

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

type WalletData struct {
	Name     string
	Accounts map[string][]string
}

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

func combineWallets(ctx context.Context) ([]WalletData, error) {
	stores, err := loadStores(ctx)
	if err != nil {
		fmt.Println(err)
	}

	accountData := make(map[string][]string)
	walletData := make([]WalletData, len(stores))

	for _, store := range stores {
		fmt.Println(store.Location)
		for _, wallet := range store.Wallets {
			fmt.Println(wallet.Name())
			for account := range wallet.Accounts(ctx) {
				key, err := getAccountKey(ctx, account)
				if err != nil {
					fmt.Println("Error")
				}
				bs, _ := json.Marshal(key)
				accountData[account.Name()] = append(
					accountData[account.Name()],
					string(bs),
				)
			}
			walletData = append(walletData,
				WalletData{
					Name:     wallet.Name(),
					Accounts: accountData,
				},
			)
		}
	}
	return walletData, nil
}
