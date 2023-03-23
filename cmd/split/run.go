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
	masterKey := service.GetMasterKey()
	accountsPasswords := service.GetAccountsPasswords()
	var peers service.Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	masterSKs, masterPKs := bls.Split(ctx, masterKey, threshold)

	store := service.CreateStore("./newwallets")
	wallet := service.CreateWallet(store)

	service.CreateAccounts(
		wallet,
		accountsPasswords[0],
		masterPKs,
		masterSKs,
		threshold,
		peers,
	)

	fmt.Println(wallet)
	//CombineStores(ctx)
	//SaveWallets(ctx)
}
