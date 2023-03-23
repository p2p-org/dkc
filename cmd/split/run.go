package split

import (
	"context"
	"fmt"

	"github.com/p2p-org/dkc/service/"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
)

func Run() {
	ctx := context.Background()
	service.CreateWallets(ctx)
	threshold := viper.GetUint32("signing-threshold")
	masterKey := getMasterKey()
	accountsPassword := getAccountsPassword()
	var peers Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	masterSKs, masterPKs := bls.Split(ctx, masterKey, threshold)

	store := service.createStore("./newwallets")
	wallet := service.createWallet(store)

	service.createAccounts(
		wallet,
		accountsPassword,
		masterPKs,
		masterSKs,
		threshold,
		peers,
	)

	fmt.Println(wallet)
	//CombineStores(ctx)
	//SaveWallets(ctx)
}
