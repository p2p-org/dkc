package split

import (
	"context"
	"encoding/hex"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/p2p-org/dkc/utils/crypto/bls"
	"github.com/spf13/viper"
	"github.com/google/uuid"
)

type Runtime struct {
	ctx            context.Context
	dWalletsPath   string
	ndWalletsPath  string
	passphrasesIn  [][]byte
	passphrasesOut [][]byte
	accountDatas   map[string]AccountExtends
	peers          utils.Peers
	threshold      uint32
	walletsMap     map[uint64]utils.DWallet
	peersIDs       []uint64
}

type AccountExtends struct {
	InitialSignature []byte
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []utils.Account
	MasterPKs        [][]byte
}

func newRuntime() (*Runtime, error) {
	rt := &Runtime{}
	var err error

	utils.LogSplit.Debug().Msg("validating nd-wallets config field")
	var ndWalletConfig utils.NDWalletConfig
	err = viper.UnmarshalKey("nd-wallets", &ndWalletConfig)
	if err != nil {
		return nil, err
	}

	err = ndWalletConfig.Validate()
	if err != nil {
		return nil, err
	}

	utils.LogSplit.Debug().Msg("validating d-wallets config field")
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
	rt.threshold = dWalletConfig.Threshold
	utils.LogSplit.Debug().Msgf("getting input passwords from %s", ndWalletConfig.Passphrases)
	rt.passphrasesIn, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	utils.LogSplit.Debug().Msgf("getting input passwords from %s", dWalletConfig.Passphrases)
	rt.passphrasesOut, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	rt.accountDatas = make(map[string]AccountExtends)
	rt.walletsMap = make(map[uint64]utils.DWallet)

	rt.peers = dWalletConfig.Peers

	utils.LogSplit.Debug().Msg("generating peersIDs")
	for id := range rt.peers {
		rt.peersIDs = append(rt.peersIDs, id)
	}

	return rt, nil
}

func (rt *Runtime) validate() error {
	if rt.dWalletsPath == rt.ndWalletsPath {
		return utils.ErrorSameDirs
	}
	return nil
}

func (rt *Runtime) createWallets() error {
	walletName := uuid.New().String()
	for id, peer := range rt.peers {
		res, err := regexp.Compile(`:.*`)
		if err != nil {
			return err
		}
		utils.LogSplit.Debug().Msgf("creating store for peer: %d", id)
		storePath := rt.dWalletsPath + "/" + res.ReplaceAllString(peer, "")
		store, err := utils.CreateStore(storePath)
		if err != nil {
			return err
		}
		utils.LogSplit.Debug().Msgf("creating wallet for peer %d", id)
		wallet, err := utils.CreateDWallet(store, walletName)
		if err != nil {
			return err
		}
		rt.walletsMap[id] = wallet
	}
	return nil
}

func (rt *Runtime) loadWallets() error {
	utils.LogSplit.Debug().Msgf("load store %s", rt.ndWalletsPath)
	s, err := utils.LoadStore(rt.ctx, rt.ndWalletsPath, rt.passphrasesIn)
	if err != nil {
		return err
	}

	for _, w := range s.Wallets {
		utils.LogSplit.Debug().Msgf("load wallet %s ", w.Name())
		for account := range w.Accounts(rt.ctx) {
			utils.LogSplit.Debug().Msgf("get private key for account %s ", account.Name())
			key, err := utils.GetAccountKey(rt.ctx, account, rt.passphrasesIn)
			if err != nil {
				return err
			}
			utils.LogSplit.Debug().Msgf("get pub key for account %s ", account.Name())
			pubKey, err := utils.GetAccountPubkey(account)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("signing test string for account %s ", account.Name())
			initialSignature, err := utils.AccountSign(rt.ctx, account, rt.passphrasesIn)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("bls split key for account %s ", account.Name())
			masterSKs, masterPKs, err := bls.Split(rt.ctx, key, rt.threshold)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("setup bls participants for account %s ", account.Name())
			participants, err := bls.SetupParticipants(masterSKs, masterPKs, rt.peersIDs, len(rt.peers))
			if err != nil {
				return err
			}

			rt.accountDatas[account.Name()] = AccountExtends{
				MasterPKs:        masterPKs,
				InitialSignature: initialSignature,
				Accounts:         participants,
				PubKey:           pubKey,
			}
		}
	}

	return nil
}

func (rt *Runtime) saveAccounts() error {
	for accountName, account := range rt.accountDatas {
		utils.LogSplit.Debug().Msgf("saving account %s ", accountName)
		for i, acc := range account.Accounts {
			utils.LogSplit.Debug().Msgf("creating account with id %d ", acc.ID)
			finalAccount, err := utils.CreateDAccount(
				rt.walletsMap[acc.ID],
				accountName,
				account.MasterPKs,
				acc.Key,
				rt.threshold,
				rt.peers,
				rt.passphrasesOut[0],
			)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("generating signature for account with id %d ", acc.ID)
			account.Accounts[i].Signature, err = utils.AccountSign(rt.ctx, finalAccount, rt.passphrasesOut)
			if err != nil {
				return err
			}
			utils.LogSplit.Debug().Msgf("getting composite pub key for account with id %d ", acc.ID)
			compositePubKey, err := utils.GetAccountCompositePubkey(finalAccount)
			if err != nil {
				return err
			}
			account.CompositePubKeys = append(account.CompositePubKeys, compositePubKey)
		}
	}

	return nil
}

func (rt *Runtime) checkSignature() error {
	for _, account := range rt.accountDatas {
		utils.LogSplit.Debug().Msgf("generating bls signature for pub key %s", hex.EncodeToString(account.PubKey))
		finalSignature, err := bls.Sign(rt.ctx, account.Accounts)
		if err != nil {
			return err
		}

		utils.LogSplit.Debug().Msgf("compare bls signatures for pub key %s", hex.EncodeToString(account.PubKey))
		err = bls.SignatureCompare(finalSignature, account.InitialSignature)
		if err != nil {
			return err
		}

		utils.LogSplit.Debug().Msgf("compare composite pubkeys for pub key %s", hex.EncodeToString(account.PubKey))
		for _, compositePubKey := range account.CompositePubKeys {
			err = bls.CompositeKeysCompare(compositePubKey, account.PubKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
