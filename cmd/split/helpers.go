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

type SplitRuntime struct {
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
	walletName     string
}

type AccountExtends struct {
	InitialSignature []byte
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []utils.Account
	MasterPKs        [][]byte
}

func newSplitRuntime() (*SplitRuntime, error) {
	sr := &SplitRuntime{}
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

	sr.ctx = context.Background()
	sr.dWalletsPath = dWalletConfig.Path
	sr.ndWalletsPath = ndWalletConfig.Path
	sr.threshold = dWalletConfig.Threshold
	sr.walletName = dWalletConfig.WalletName
	utils.LogSplit.Debug().Msgf("getting input passwords from %s", ndWalletConfig.Passphrases)
	sr.passphrasesIn, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	utils.LogSplit.Debug().Msgf("getting input passwords from %s", dWalletConfig.Passphrases)
	sr.passphrasesOut, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	sr.accountDatas = make(map[string]AccountExtends)
	sr.walletsMap = make(map[uint64]utils.DWallet)

	sr.peers = dWalletConfig.Peers

	utils.LogSplit.Debug().Msg("generating peersIDs")
	for id := range sr.peers {
		sr.peersIDs = append(sr.peersIDs, id)
	}

	return sr, nil
}

func (sr *SplitRuntime) validate() error {
	if sr.dWalletsPath == sr.ndWalletsPath {
		return utils.ErrorSameDirs
	}
	return nil
}

func (sr *SplitRuntime) createWallets() error {
	walletName := uuid.New().String()
	if sr.walletName != "" {
		walletName = sr.walletName
	}
	for id, peer := range sr.peers {
		res, err := regexp.Compile(`:.*`)
		if err != nil {
			return err
		}
		utils.LogSplit.Debug().Msgf("creating store for peer: %d", id)
		storePath := sr.dWalletsPath + "/" + res.ReplaceAllString(peer, "")
		store, err := utils.CreateStore(storePath)
		if err != nil {
			return err
		}
		utils.LogSplit.Debug().Msgf("creating wallet for peer %d", id)
		wallet, err := utils.CreateDWallet(store, walletName)
		if err != nil {
			return err
		}
		sr.walletsMap[id] = wallet
	}
	return nil
}

func (sr *SplitRuntime) loadWallets() error {
	utils.LogSplit.Debug().Msgf("load store %s", sr.ndWalletsPath)
	s, err := utils.LoadStore(sr.ctx, sr.ndWalletsPath, sr.passphrasesIn)
	if err != nil {
		return err
	}

	for _, w := range s.Wallets {
		utils.LogSplit.Debug().Msgf("load wallet %s ", w.Name())
		for account := range w.Accounts(sr.ctx) {
			utils.LogSplit.Debug().Msgf("get private key for account %s ", account.Name())
			key, err := utils.GetAccountKey(sr.ctx, account, sr.passphrasesIn)
			if err != nil {
				return err
			}
			utils.LogSplit.Debug().Msgf("get pub key for account %s ", account.Name())
			pubKey, err := utils.GetAccountPubkey(account)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("signing test string for account %s ", account.Name())
			initialSignature, err := utils.AccountSign(sr.ctx, account, sr.passphrasesIn)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("bls split key for account %s ", account.Name())
			masterSKs, masterPKs, err := bls.Split(sr.ctx, key, sr.threshold)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("setup bls participants for account %s ", account.Name())
			participants, err := bls.SetupParticipants(masterSKs, masterPKs, sr.peersIDs, len(sr.peers))
			if err != nil {
				return err
			}

			sr.accountDatas[account.Name()] = AccountExtends{
				MasterPKs:        masterPKs,
				InitialSignature: initialSignature,
				Accounts:         participants,
				PubKey:           pubKey,
			}
		}
	}

	return nil
}

func (sr *SplitRuntime) saveAccounts() error {
	for accountName, account := range sr.accountDatas {
		utils.LogSplit.Debug().Msgf("saving account %s ", accountName)
		for i, acc := range account.Accounts {
			utils.LogSplit.Debug().Msgf("creating account with id %d ", acc.ID)
			finalAccount, err := utils.CreateDAccount(
				sr.walletsMap[acc.ID],
				accountName,
				account.MasterPKs,
				acc.Key,
				sr.threshold,
				sr.peers,
				sr.passphrasesOut[0],
			)
			if err != nil {
				return err
			}

			utils.LogSplit.Debug().Msgf("generating signature for account with id %d ", acc.ID)
			account.Accounts[i].Signature, err = utils.AccountSign(sr.ctx, finalAccount, sr.passphrasesOut)
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

func (sr *SplitRuntime) checkSignature() error {
	for _, account := range sr.accountDatas {
		utils.LogSplit.Debug().Msgf("generating bls signature for pub key %s", hex.EncodeToString(account.PubKey))
		finalSignature, err := bls.Sign(sr.ctx, account.Accounts)
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
