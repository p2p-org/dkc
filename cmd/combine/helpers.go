package combine

import (
	"context"
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
	ctx            context.Context
	dWalletsPath   string
	ndWalletsPath  string
	passphrasesIn  [][]byte
	passphrasesOut [][]byte
	accountDatas   map[string]AccountExtends
	stores         []utils.DirkStore
	peers          utils.Peers
	wallet         utils.NDWallet
	store          types.Store
}

func newCombineRuntime() (*CombineRuntime, error) {
	cr := &CombineRuntime{}
	var err error

	var ndWalletConfig utils.NDWalletConfig
	err = viper.UnmarshalKey("nd-wallets", &ndWalletConfig)
	if err != nil {
		return nil, err
	}

	var dWalletConfig utils.DWalletConfig
	err = viper.UnmarshalKey("distributed-wallets", &dWalletConfig)
	if err != nil {
		return nil, err
	}

	cr.ctx = context.Background()
	cr.dWalletsPath = dWalletConfig.Path
	cr.ndWalletsPath = ndWalletConfig.Path
	cr.passphrasesIn, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	cr.passphrasesOut, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	cr.accountDatas = make(map[string]AccountExtends)
	cr.stores, err = utils.LoadStores(cr.ctx, cr.dWalletsPath, cr.passphrasesIn)
	if err != nil {
		return nil, err
	}

	cr.peers = dWalletConfig.Peers

	return cr, nil
}

func (cr *CombineRuntime) createWalletAndStore() error {
	var err error
	cr.store, err = utils.CreateStore(cr.ndWalletsPath)
	if err != nil {
		return err
	}
	cr.wallet, err = utils.CreateNDWallet(cr.store)
	if err != nil {
		return err
	}
	return nil
}

func (cr *CombineRuntime) checkSignature() error {
	for accountName, account := range cr.accountDatas {
		key, err := bls.Recover(cr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		initialSignature, err := bls.Sign(cr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		finalAccount, err := utils.CreateNDAccount(cr.wallet, accountName, key, cr.passphrasesOut[0])
		if err != nil {
			return err
		}

		finalSignature, err := utils.AccountSign(cr.ctx, finalAccount, cr.passphrasesOut)
		if err != nil {
			return err
		}

		err = bls.SignatureCompare(finalSignature, initialSignature)
		if err != nil {
			return err
		}

		pubkey, err := utils.GetAccountPubkey(finalAccount)
		if err != nil {
			return err
		}

		for _, compositeKey := range account.CompositePubKeys {
			err = bls.CompositeKeysCompare(compositeKey, pubkey)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (cr *CombineRuntime) storeUpdater() error {
	for _, store := range cr.stores {
		var participantID uint64
		for id := range cr.peers {
			peerExists, err := regexp.MatchString(filepath.Base(store.Location)+":.*", cr.peers[id])
			if err != nil {
				return err
			}
			if !peerExists {
				continue
			}
			participantID = id

			for _, wallet := range store.Wallets {
				for account := range wallet.Accounts(cr.ctx) {
					key, err := utils.GetAccountKey(cr.ctx, account, cr.passphrasesOut)
					if err != nil {
						return err
					}

					initialSignature, err := utils.AccountSign(cr.ctx, account, cr.passphrasesOut)
					if err != nil {
						return err
					}
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
