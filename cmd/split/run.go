package split

import (
	"context"
	"fmt"

	"github.com/p2p-org/dkc/service"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
)

func Run() {
	ctx := context.Background()
	threshold := viper.GetUint32("signing-threshold")
	accountsPasswords := service.GetAccountsPasswords()
	var peers service.Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	stores, err := service.LoadStores(ctx, "./restoredwallets", accountsPasswords)
	if err != nil {
		fmt.Println(err)
	}

	store := service.CreateStore("./newwallets")
	wallet := service.CreateWallet(store, "distributed")

	for _, s := range stores {
		for _, wallet := range s.Wallets {
			for account := range wallet.Accounts(ctx) {
				key, err := service.GetAccountKey(ctx, account, accountsPasswords)
				if err != nil {
					fmt.Println("Error")
				}

				masterSKs, masterPKs := bls.Split(ctx, key, threshold)
				service.CreateAccounts(
					wallet,
					accountsPasswords[0],
					account.Name(),
					masterPKs,
					masterSKs,
					threshold,
					peers,
				)
			}
		}
	}

	fmt.Println(wallet)
}
