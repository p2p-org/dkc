package combine

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/p2p-org/dkc/service"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
)

func Run() {
	ctx := context.Background()
	signString := "testingStringABC"
	walletDir := viper.GetString("walletDir")
	var peers service.Peers
	passphrases := service.GetAccountsPasswords()
	accountDatas := make(map[string][]service.Account)
	stores, err := service.LoadStores(ctx, walletDir, passphrases)
	if err != nil {
		fmt.Println(err)
	}

	err = viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	for _, store := range stores {
		var participantID uint64
		for id := range peers {
			peerExists, _ := regexp.MatchString(filepath.Base(store.Location)+":.*", peers[id])
			if peerExists {
				participantID = id

				for _, wallet := range store.Wallets {
					for account := range wallet.Accounts(ctx) {
						key, err := service.GetAccountKey(ctx, account, passphrases)
						if err != nil {
							fmt.Println("Error")
						}

						initialSignature := service.AccountSign(ctx, account, []byte(signString), passphrases)

						accountDatas[account.Name()] = append(
							accountDatas[account.Name()],
							service.Account{
								Key:       key,
								Signature: initialSignature,
								ID:        participantID,
							},
						)
					}
				}
			}
		}
	}

	store := service.CreateStore("./restoredwallets")
	wallet := service.CreateWallet(store, "non-deterministic")
	for accountName, account := range accountDatas {
		key, err := bls.Recover(ctx, account)
		if err != nil {
			panic(err)
		}

		initialSignature := bls.Sign(ctx, account)

		finalAccount := service.CreateNDAccount(key, accountName, passphrases[0], wallet)

		finalSignature := service.AccountSign(ctx, finalAccount, []byte(signString), passphrases)

		if !bytes.Equal(finalSignature, initialSignature) {
			panic("test")
		}
		fmt.Println(finalAccount)
	}

	return
}
