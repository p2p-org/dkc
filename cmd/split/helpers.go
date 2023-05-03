package split

import (
	"bytes"
	"context"
	"os"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/p2p-org/dkc/utils/crypto/bls"
	"github.com/spf13/viper"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type SplitRuntime struct {
	ctx                    context.Context
	distributedWalletsPath string
	ndWalletsPath          string
	passphrases            [][]byte
	accountDatas           map[string]AccountExtends
	peers                  utils.Peers
	threshold              uint32
	walletsMap             map[uint64]types.Wallet
	peersIDs               []uint64
}

type AccountExtends struct {
	InitialSignature []byte
	PubKey           []byte
	CompositePubKeys [][]byte
	Accounts         []utils.Account
	MasterPKs        [][]byte
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

func newSplitRuntime() (*SplitRuntime, error) {
	var peers utils.Peers
	sr := &SplitRuntime{}
	var err error

	sr.ctx = context.Background()
	sr.distributedWalletsPath = viper.GetString("distributed-wallets")
	sr.ndWalletsPath = viper.GetString("nd-wallets")
	sr.threshold = viper.GetUint32("signing-threshold")
	sr.passphrases = getAccountsPasswords()
	sr.accountDatas = make(map[string]AccountExtends)
	sr.walletsMap = make(map[uint64]types.Wallet)

	err = viper.UnmarshalKey("peers", &peers)
	if err != nil {
		return sr, err
	}
	sr.peers = peers

	return sr, nil
}

func (sr *SplitRuntime) createWallets() error {
	var peersIDs []uint64
	for id, peer := range sr.peers {
		peersIDs = append(peersIDs, id)
		res := regexp.MustCompile(`:.*`)
		storePath := sr.distributedWalletsPath + "/" + res.ReplaceAllString(peer, "")
		store := utils.CreateStore(storePath)
		wallet := utils.CreateWallet(store, "distributed")
		sr.walletsMap[id] = wallet
	}
	sr.peersIDs = peersIDs
	return nil
}

func (sr *SplitRuntime) loadWallets() error {
	s, err := utils.LoadStore(sr.ctx, sr.ndWalletsPath, sr.passphrases)
	if err != nil {
		return err
	}

	for _, w := range s.Wallets {
		for account := range w.Accounts(sr.ctx) {
			key, err := utils.GetAccountKey(sr.ctx, account, sr.passphrases)
			if err != nil {
				return err
			}
			pubKey, err := utils.GetAccountPubkey(account)
			if err != nil {
				return err
			}

			initialSignature := utils.AccountSign(sr.ctx, account, sr.passphrases)

			masterSKs, masterPKs := bls.Split(sr.ctx, key, sr.threshold)

			sr.accountDatas[account.Name()] = AccountExtends{
				MasterPKs:        masterPKs,
				InitialSignature: initialSignature,
				Accounts:         bls.SetupParticipants(masterSKs, masterPKs, sr.peersIDs, len(sr.peers)),
				PubKey:           pubKey,
			}
		}
	}

	return nil
}

func (sr *SplitRuntime) saveAccounts() error {
	for accountName, account := range sr.accountDatas {
		for i, acc := range account.Accounts {
			finalAccount := utils.CreateAccount(
				sr.walletsMap[acc.ID],
				accountName,
				account.MasterPKs,
				acc.Key,
				sr.threshold,
				sr.peers,
				sr.passphrases[0],
			)

			account.Accounts[i].Signature = utils.AccountSign(sr.ctx, finalAccount, sr.passphrases)
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
		finalSignature := bls.Sign(sr.ctx, account.Accounts)
		if !bytes.Equal(finalSignature, account.InitialSignature) {
			panic("test")
		}

		for _, compositePubKey := range account.CompositePubKeys {
			if !bytes.Equal(compositePubKey, account.PubKey) {
				panic("test")
			}
		}
	}

	return nil
}
