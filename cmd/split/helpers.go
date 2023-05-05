package split

import (
	"context"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/p2p-org/dkc/utils/crypto/bls"
	"github.com/spf13/viper"
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

	sr.ctx = context.Background()
	sr.dWalletsPath = dWalletConfig.Path
	sr.ndWalletsPath = ndWalletConfig.Path
	sr.threshold = dWalletConfig.Threshold
	sr.passphrasesIn, err = utils.GetAccountsPasswords(ndWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	sr.passphrasesOut, err = utils.GetAccountsPasswords(dWalletConfig.Passphrases)
	if err != nil {
		return nil, err
	}
	sr.accountDatas = make(map[string]AccountExtends)
	sr.walletsMap = make(map[uint64]utils.DWallet)

	sr.peers = dWalletConfig.Peers

	return sr, nil
}

func (sr *SplitRuntime) createWallets() error {
	var peersIDs []uint64
	for id, peer := range sr.peers {
		peersIDs = append(peersIDs, id)
		res, err := regexp.Compile(`:.*`)
		if err != nil {
			return err
		}
		storePath := sr.dWalletsPath + "/" + res.ReplaceAllString(peer, "")
		store, err := utils.CreateStore(storePath)
		if err != nil {
			return err
		}
		wallet, err := utils.CreateDWallet(store)
		if err != nil {
			return err
		}
		sr.walletsMap[id] = wallet
	}
	sr.peersIDs = peersIDs
	return nil
}

func (sr *SplitRuntime) loadWallets() error {
	s, err := utils.LoadStore(sr.ctx, sr.ndWalletsPath, sr.passphrasesIn)
	if err != nil {
		return err
	}

	for _, w := range s.Wallets {
		for account := range w.Accounts(sr.ctx) {
			key, err := utils.GetAccountKey(sr.ctx, account, sr.passphrasesIn)
			if err != nil {
				return err
			}
			pubKey, err := utils.GetAccountPubkey(account)
			if err != nil {
				return err
			}

			initialSignature, err := utils.AccountSign(sr.ctx, account, sr.passphrasesIn)
			if err != nil {
				return err
			}

			masterSKs, masterPKs, err := bls.Split(sr.ctx, key, sr.threshold)
			if err != nil {
				return err
			}

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
		for i, acc := range account.Accounts {
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

			account.Accounts[i].Signature, err = utils.AccountSign(sr.ctx, finalAccount, sr.passphrasesOut)
			if err != nil {
				return err
			}
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
		finalSignature, err := bls.Sign(sr.ctx, account.Accounts)
		if err != nil {
			return err
		}

		err = bls.SignatureCompare(finalSignature, account.InitialSignature)
		if err != nil {
			return err
		}

		for _, compositePubKey := range account.CompositePubKeys {
			err = bls.CompositeKeysCompare(compositePubKey, account.PubKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
