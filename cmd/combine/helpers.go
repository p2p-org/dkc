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

type Runtime struct {
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

func newRuntime() (*Runtime, error) {
	rt := &Runtime{}
	var err error

	utils.LogCombine.Debug().Msg("validating nd-wallets field")
	var ndWalletConfig utils.NDWalletConfig
	err = viper.UnmarshalKey("nd-wallets", &ndWalletConfig)
	if err != nil {
		return nil, err
	}

	err = ndWalletConfig.Validate()
	if err != nil {
		return nil, err
	}

	utils.LogCombine.Debug().Msg("validating distributed-wallets field")
	var dWalletConfig utils.DWalletConfig
	err = viper.UnmarshalKey("distributed-wallets", &dWalletConfig)
	if err != nil {
		return nil, err
	}

	err = dWalletConfig.Validate()
	if err != nil {
		return nil, err
	}

	rt.ctx = context.Background()
	rt.dWalletsPath = dWalletConfig.Path
	rt.ndWalletsPath = ndWalletConfig.Path
	utils.LogCombine.Debug().Msgf("getting input passwords form file %s", dWalletConfig.Passphrases)
	rt.passphrasesIn, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	utils.LogCombine.Debug().Msgf("getting output passwords form file %s", ndWalletConfig.Passphrases)
	rt.passphrasesOut, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	rt.accountDatas = make(map[string]AccountExtends)
	utils.LogCombine.Debug().Msgf("loading stores form %s", rt.dWalletsPath)
	rt.stores, err = utils.LoadStores(rt.ctx, rt.dWalletsPath, rt.passphrasesIn)
	if err != nil {
		return nil, err
	}

	rt.peers = dWalletConfig.Peers

	return rt, nil
}

func (rt *Runtime) validate() error {
	if rt.dWalletsPath == rt.ndWalletsPath {
		return utils.ErrorSameDirs
	}
	return nil
}

func (rt *Runtime) createWalletAndStore() error {
	var err error
	utils.LogCombine.Debug().Msgf("creating store %s", rt.ndWalletsPath)
	rt.store, err = utils.CreateStore(rt.ndWalletsPath)
	if err != nil {
		return err
	}
	utils.LogCombine.Debug().Msg("creating ndwallet")
	rt.wallet, err = utils.CreateNDWallet(rt.store)
	if err != nil {
		return err
	}
	return nil
}

func (rt *Runtime) checkSignature() error {
	for accountName, account := range rt.accountDatas {
		utils.LogCombine.Debug().Msgf("recover private key for account %s", accountName)
		key, err := bls.Recover(rt.ctx, account.Accounts)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("bls sing for account %s", accountName)
		initialSignature, err := bls.Sign(rt.ctx, account.Accounts)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("creating nd account for account %s", accountName)
		finalAccount, err := utils.CreateNDAccount(rt.wallet, accountName, key, rt.passphrasesOut[0])
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("signing message for account %s", accountName)
		finalSignature, err := utils.AccountSign(rt.ctx, finalAccount, rt.passphrasesOut)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("comapare signatures for account %s", accountName)
		err = bls.SignatureCompare(finalSignature, initialSignature)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("get pub key for account %s", accountName)
		pubkey, err := utils.GetAccountPubkey(finalAccount)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("get composite pub key for account %s", accountName)
		for _, compositeKey := range account.CompositePubKeys {
			err = bls.CompositeKeysCompare(compositeKey, pubkey)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (rt *Runtime) storeUpdater() error {
	for _, store := range rt.stores {
		var participantID uint64
		for id := range rt.peers {
			peerExists, err := regexp.MatchString(filepath.Base(store.Location)+":.*", rt.peers[id])
			if err != nil {
				return err
			}
			if !peerExists {
				continue
			}
			participantID = id

			for _, wallet := range store.Wallets {
				utils.LogCombine.Debug().Msgf("loading data for wallet %s", wallet.Name())
				for account := range wallet.Accounts(rt.ctx) {
					utils.LogCombine.Debug().Msgf("get private key for account %s", account.Name())
					key, err := utils.GetAccountKey(rt.ctx, account, rt.passphrasesOut)
					if err != nil {
						return err
					}

					utils.LogCombine.Debug().Msgf("sign message from account %s", account.Name())
					initialSignature, err := utils.AccountSign(rt.ctx, account, rt.passphrasesOut)
					if err != nil {
						return err
					}
					utils.LogCombine.Debug().Msgf("get composite pub key for account %s", account.Name())
					compositePubKey, err := utils.GetAccountCompositePubkey(account)
					if err != nil {
						return err
					}

					rt.accountDatas[account.Name()] = AccountExtends{
						Accounts: append(rt.accountDatas[account.Name()].Accounts,
							utils.Account{
								Key:       key,
								Signature: initialSignature,
								ID:        participantID,
							},
						),
						CompositePubKeys: append(rt.accountDatas[account.Name()].CompositePubKeys, compositePubKey),
					}
				}
			}
		}
	}
	return nil
}
