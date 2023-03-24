package split

import (
	"bytes"
	"context"
	"fmt"

	"github.com/p2p-org/dkc/service"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
)

func Run() {
	ctx := context.Background()
	signString := "testingStringABC"
	threshold := viper.GetUint32("signing-threshold")
	accountsPasswords := service.GetAccountsPasswords()
	var peers service.Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	s, err := service.LoadStore(ctx, "./restoredwallets", accountsPasswords)
	if err != nil {
		fmt.Println(err)
	}

	store := service.CreateStore("./newwallets")
	wallet := service.CreateWallet(store, "distributed")

	for _, w := range s.Wallets {
		for account := range w.Accounts(ctx) {
			key, err := service.GetAccountKey(ctx, account, accountsPasswords)
			if err != nil {
				fmt.Println("Error")
			}

			masterSKs, masterPKs := bls.Split(ctx, key, threshold)
			initialSignature := service.AccountSign(ctx, account, []byte(signString), accountsPasswords)

			finalAccount := service.CreateAccount(
				wallet,
				accountsPasswords[0],
				account.Name(),
				masterPKs,
				masterSKs,
				threshold,
				peers,
			)

			finalSignature := service.AccountSign(ctx, finalAccount, []byte(signString), accountsPasswords)
			fmt.Println(initialSignature)
			fmt.Println(finalSignature)

			if !bytes.Equal(finalSignature, initialSignature) {
				panic("test")
			}
		}
	}

	return
}
