package store

import (
	"context"
	"fmt"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type HDStore struct {
	Type        string
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
}

func (s *HDStore) Create() error {
	_, err := createStore(s.Path)
	if err != nil {
		return err
	}

	return nil
}
func (s *HDStore) GetWalletsAccountsMap() ([]AccountsData, []string, error) {
	a, w, err := getWalletsAccountsMap(s.Ctx, s.Path)
	if err != nil {
		return nil, nil, err
	}

	return a, w, nil
}

func (s *HDStore) CreateWallet(name string) (types.Wallet, error) {
	store, err := getStore(s.Path)
	if err != nil {
		return nil, err
	}
	wallet, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(store))
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *HDStore) GetPK(w string, a string) ([]byte, error) {
	wallet, err := getWallet(s.Path, w)
	if err != nil {
		return nil, err
	}
	account, err := wallet.(types.WalletAccountByNameProvider).AccountByName(s.Ctx, a)
	if err != nil {
		return nil, err
	}

	key, err := getAccountPK(account, s.Ctx, s.Passphrases)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (s *HDStore) GetPath() string {
	return s.Path
}

func (s *HDStore) GetType() string {
	return s.Type
}

func newHDStore(t string) (HDStore, error) {
	s := HDStore{}
	//Parse Wallet Type
	wt := viper.GetString(fmt.Sprintf("%s.wallet.type", t))
	utils.Log.Debug().Msgf("setting store type to %s", wt)
	s.Type = wt

	//Parse Store Path
	storePath := viper.GetString(fmt.Sprintf("%s.store.path", t))
	utils.Log.Debug().Msgf("setting store path to %s", storePath)
	if storePath == "" {
		return s, errors.New("hd store path is empty")
	}
	s.Path = storePath

	//Parse Passphrases
	utils.Log.Debug().Msgf("getting passhphrases")
	passphrases, err := getAccountsPasswords(
		viper.GetString(fmt.Sprintf("%s.wallet.passphrases.path", t)),
	)
	if err != nil {
		return s, err
	}
	utils.Log.Debug().Msgf("checking passhphrases len: %d", len(passphrases))
	if len(passphrases) == 0 {
		return s, errors.New("passhparases file for hd store is empty")
	}

	// Cheking If Passphrases Index Is Set
	if viper.IsSet(fmt.Sprintf("%s.wallet.passphrases.index", t)) {
		index := viper.GetInt(fmt.Sprintf("%s.wallet.passphrases.index", t))
		passphrases = [][]byte{passphrases[index]}
	}

	s.Passphrases = passphrases

	return s, nil
}
