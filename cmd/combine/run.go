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

type AccountExtends struct {
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []service.Account
}

func Run() {
	ctx := context.Background()
	signString := "testingStringABC"
	walletDir := viper.GetString("walletDir")
	var peers service.Peers
	passphrases := service.GetAccountsPasswords()
	accountDatas := make(map[string]AccountExtends)
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
						compositePubKey, err := service.GetAccountCompositePubkey(account)
						if err != nil {
							panic(err)
						}

						accountDatas[account.Name()] = AccountExtends{
							Accounts: append(accountDatas[account.Name()].Accounts,
								service.Account{
									Key:       key,
									Signature: initialSignature,
									ID:        participantID,
								},
							),
							CompositePubKeys: append(accountDatas[account.Name()].CompositePubKeys, compositePubKey),
						}
					}
				}
			}
		}
	}

	store := service.CreateStore("./restoredwallets")
	wallet := service.CreateWallet(store, "non-deterministic")
	for accountName, account := range accountDatas {
		key, err := bls.Recover(ctx, account.Accounts)
		if err != nil {
			panic(err)
		}

		initialSignature := bls.Sign(ctx, account.Accounts)

		finalAccount := service.CreateNDAccount(key, accountName, passphrases[0], wallet)

		finalSignature := service.AccountSign(ctx, finalAccount, []byte(signString), passphrases)

		if !bytes.Equal(finalSignature, initialSignature) {
			panic("test")
		}

		pubkey, err := service.GetAccountPubkey(finalAccount)
		if err != nil {
			panic(err)
		}

		for _, compositeKey := range account.CompositePubKeys {
			if !bytes.Equal(compositeKey, pubkey) {
				panic("test")
			}

		}
	}

	return
}
