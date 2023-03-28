package split

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/p2p-org/dkc/service"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type AccountExtends struct {
	InitialSignature []byte
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []service.Account
	MasterPKs        [][]byte
}

func Run() {
	ctx := context.Background()
	signString := "testingStringABC"
	threshold := viper.GetUint32("signing-threshold")
	accountsPasswords := service.GetAccountsPasswords()
	accountDatas := make(map[string]AccountExtends)
	walletsMap := make(map[uint64]types.Wallet)
	var peers service.Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}
	var peersIDs []uint64
	for id, peer := range peers {
		peersIDs = append(peersIDs, id)
		res := regexp.MustCompile(`:.*`)
		storePath := "./newwallets/" + res.ReplaceAllString(peer, "")
		store := service.CreateStore(storePath)
		wallet := service.CreateWallet(store, "distributed")
		walletsMap[id] = wallet
	}

	s, err := service.LoadStore(ctx, "./restoredwallets", accountsPasswords)
	if err != nil {
		fmt.Println(err)
	}

	for _, w := range s.Wallets {
		for account := range w.Accounts(ctx) {
			key, err := service.GetAccountKey(ctx, account, accountsPasswords)
			if err != nil {
				fmt.Println("Error")
			}
			pubKey, err := service.GetAccountPubkey(account)
			if err != nil {
				panic(err)
			}

			initialSignature := service.AccountSign(ctx, account, []byte(signString), accountsPasswords)

			masterSKs, masterPKs := bls.Split(ctx, key, threshold)

			accountDatas[account.Name()] = AccountExtends{
				MasterPKs:        masterPKs,
				InitialSignature: initialSignature,
				Accounts:         bls.SetupParticipants(masterSKs, masterPKs, peersIDs, len(peers)),
				PubKey:           pubKey,
			}
		}
	}

	for accountName, account := range accountDatas {
		for i, acc := range account.Accounts {
			finalAccount := service.CreateAccount(
				walletsMap[acc.ID],
				accountsPasswords[0],
				accountName,
				account.MasterPKs,
				acc.Key,
				threshold,
				peers,
			)
			accountDatas[accountName].Accounts[i].Signature = service.AccountSign(ctx, finalAccount, []byte(signString), accountsPasswords)
			compositePubKey, err := service.GetAccountCompositePubkey(finalAccount)
			if err != nil {
				panic(err)
			}
			account.CompositePubKeys = append(account.CompositePubKeys, compositePubKey)
		}
	}

	for _, account := range accountDatas {
		finalSignature := bls.Sign(ctx, account.Accounts)
		if !bytes.Equal(finalSignature, account.InitialSignature) {
			panic("test")
		}

		for _, compositePubKey := range account.CompositePubKeys {
			if !bytes.Equal(compositePubKey, account.PubKey) {
				panic("test")
			}
		}
	}

	return
}
