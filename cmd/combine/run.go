package combine

import (
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
	walletDir := viper.GetString("walletDir")
	var peers service.Peers
	passphrases := service.GetAccountsPasswords()
	participantsIDs := make([]uint64, 0)
	accountDatas := make(service.Accounts)
	stores, err := service.LoadStores(ctx, walletDir, passphrases)
	if err != nil {
		fmt.Println(err)
	}

	err = viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	for _, store := range stores {
		for id := range peers {
			peerExists, _ := regexp.MatchString(filepath.Base(store.Location)+":.*", peers[id])
			if peerExists {
				participantsIDs = append(participantsIDs, id)
			}
		}
		for _, wallet := range store.Wallets {
			for account := range wallet.Accounts(ctx) {
				key, err := service.GetAccountKey(ctx, account, passphrases)
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

	store := service.CreateStore("./restoredwallets")
	wallet := service.CreateWallet(store, "non-deterministic")
	for accountName, account := range accountDatas {
		key, err := bls.Recover(ctx, account, participantsIDs)
		if err != nil {
			panic(err)
		}

		service.CreateNDAccount(key, accountName, passphrases[0], wallet)
	}

	return
}
