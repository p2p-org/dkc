package combine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/p2p-org/dkc/utils/crypto/bls"
	"github.com/spf13/viper"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type AccountExtends struct {
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []utils.Account
}

type CombineRuntime struct {
	ctx                    context.Context
	distributedWalletsPath string
	ndWalletsPath          string
	passphrases            [][]byte
	accountDatas           map[string]AccountExtends
	stores                 []utils.DirkStore
	peers                  utils.Peers
	wallet                 utils.NDWallet
	store                  types.Store
}

func getAccountsPasswords() [][]byte {
	accountsPasswordPath := viper.GetString("passphrases")

	content, err := os.ReadFile(accountsPasswordPath)
	if err != nil {
		panic(err)
	}

	accountsPasswords := bytes.Split(content, []byte{'\n'})
	return accountsPasswords
}

func newCombineRuntime() (*CombineRuntime, error) {
	var peers utils.Peers
	cr := &CombineRuntime{}
	var err error

	cr.ctx = context.Background()
	cr.distributedWalletsPath = viper.GetString("distributed-wallets")
	cr.ndWalletsPath = viper.GetString("nd-wallets")
	cr.passphrases = getAccountsPasswords()
	cr.accountDatas = make(map[string]AccountExtends)
	cr.stores, err = utils.LoadStores(cr.ctx, cr.distributedWalletsPath, cr.passphrases)
	if err != nil {
		return &CombineRuntime{}, err
	}

	err = viper.UnmarshalKey("peers", &peers)
	if err != nil {
		return &CombineRuntime{}, err
	}

	cr.peers = peers

	return cr, nil
}

func (cr *CombineRuntime) createWalletAndStore() error {
	cr.store = utils.CreateStore(cr.ndWalletsPath)
	cr.wallet = utils.CreateNDWallet(cr.store)
	return nil
}

func (cr *CombineRuntime) checkSignature() error {
	for accountName, account := range cr.accountDatas {
		key, err := bls.Recover(cr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		initialSignature := bls.Sign(cr.ctx, account.Accounts)

		finalAccount := utils.CreateNDAccount(cr.wallet, accountName, key, cr.passphrases[0])

		finalSignature := utils.AccountSign(cr.ctx, finalAccount, cr.passphrases)

		if !bytes.Equal(finalSignature, initialSignature) {
			return err
		}

		pubkey, err := utils.GetAccountPubkey(finalAccount)
		if err != nil {
			return err
		}

		for _, compositeKey := range account.CompositePubKeys {
			if !bytes.Equal(compositeKey, pubkey) {
				panic("test")
			}

		}
	}
	return nil
}

func (cr *CombineRuntime) storeUpdater() error {
	for _, store := range cr.stores {
		var participantID uint64
		for id := range cr.peers {
			peerExists, _ := regexp.MatchString(filepath.Base(store.Location)+":.*", cr.peers[id])
			if !peerExists {
				continue
			}
			participantID = id

			for _, wallet := range store.Wallets {
				for account := range wallet.Accounts(cr.ctx) {
					key, err := utils.GetAccountKey(cr.ctx, account, cr.passphrases)
					if err != nil {
						return err
					}

					initialSignature := utils.AccountSign(cr.ctx, account, cr.passphrases)
					compositePubKey, err := utils.GetAccountCompositePubkey(account)
					if err != nil {
						return err
					}

					cr.accountDatas[account.Name()] = AccountExtends{
						Accounts: append(cr.accountDatas[account.Name()].Accounts,
							utils.Account{
								Key:       key,
								Signature: initialSignature,
								ID:        participantID,
							},
						),
						CompositePubKeys: append(cr.accountDatas[account.Name()].CompositePubKeys, compositePubKey),
					}
				}
			}
		}
	}
	return nil
}
