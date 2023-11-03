package convert

import (
	"context"
	"path/filepath"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/p2p-org/dkc/utils/crypto/bls"
	"github.com/spf13/viper"
)

func newConvertRuntime() (*ConvertRuntime, error) {
	cr := &utils.ConvertRuntime{}
	var err error

	utils.LogConvert.Debug().Msg("validating input wallet")
	var inputW utils.W
	err = viper.UnmarshalKey("input", &inputW)
	if err != nil {
		return nil, err
	}

	err = inputW.Validate()
	if err != nil {
		return nil, err
	}

	utils.LogConvert.Debug().Msg("validating output field")
	var outputW utils.W
	err = viper.UnmarshalKey("output", &outputW)
	if err != nil {
		return nil, err
	}

	err = outputW.Validate()
	if err != nil {
		return nil, err
	}

	cr.ctx = context.Background()
	cr.dWalletsPath = dWalletConfig.Path
	cr.ndWalletsPath = ndWalletConfig.Path
	utils.LogCombine.Debug().Msgf("getting input passwords form file %s", dWalletConfig.Passphrases)
	cr.passphrasesIn, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	utils.LogCombine.Debug().Msgf("getting output passwords form file %s", ndWalletConfig.Passphrases)
	cr.passphrasesOut, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	cr.accountDatas = make(map[string]AccountExtends)
	utils.LogCombine.Debug().Msgf("loading stores form %s", cr.dWalletsPath)
	cr.stores, err = utils.LoadStores(cr.ctx, cr.dWalletsPath, cr.passphrasesIn)
	if err != nil {
		return nil, err
	}

	cr.peers = dWalletConfig.Peers

	return cr, nil
}

func (cr *utils.ConvertRuntime) validate() error {
	if cr.dWalletsPath == cr.ndWalletsPath {
		return utils.ErrorSameDirs
	}
	return nil
}

func (cr *CombineRuntime) createWalletAndStore() error {
	var err error
	utils.LogCombine.Debug().Msgf("creating store %s", cr.ndWalletsPath)
	cr.store, err = utils.CreateStore(cr.ndWalletsPath)
	if err != nil {
		return err
	}
	utils.LogCombine.Debug().Msg("creating ndwallet")
	cr.wallet, err = utils.CreateNDWallet(cr.store)
	if err != nil {
		return err
	}
	return nil
}

func (cr *CombineRuntime) checkSignature() error {
	for accountName, account := range cr.accountDatas {
		utils.LogCombine.Debug().Msgf("recover private key for account %s", accountName)
		key, err := bls.Recover(cr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("bls sing for account %s", accountName)
		initialSignature, err := bls.Sign(cr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("creating nd account for account %s", accountName)
		finalAccount, err := utils.CreateNDAccount(cr.wallet, accountName, key, cr.passphrasesOut[0])
		if err != nil {
			return err
		}

		utils.LogCombine.Debug().Msgf("signing message for account %s", accountName)
		finalSignature, err := utils.AccountSign(cr.ctx, finalAccount, cr.passphrasesOut)
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
				utils.LogCombine.Debug().Msgf("loading data for wallet %s", wallet.Name())
				for account := range wallet.Accounts(cr.ctx) {
					utils.LogCombine.Debug().Msgf("get private key for account %s", account.Name())
					key, err := utils.GetAccountKey(cr.ctx, account, cr.passphrasesOut)
					if err != nil {
						return err
					}

					utils.LogCombine.Debug().Msgf("sign message from account %s", account.Name())
					initialSignature, err := utils.AccountSign(cr.ctx, account, cr.passphrasesOut)
					if err != nil {
						return err
					}
					utils.LogCombine.Debug().Msgf("get composite pub key for account %s", account.Name())
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
