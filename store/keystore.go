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

type KeystoreStore struct {
	Type        string
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
}

func (s *KeystoreStore) Create() error {
	_, err := createStore(s.Path)
	if err != nil {
		return err
	}

	return nil
}

func (s *KeystoreStore) GetWalletsAccountsMap() ([]AccountsData, []string, error) {
	a, w, err := getWalletsAccountsMap(s.Ctx, s.Path)
	if err != nil {
		return nil, nil, err
	}

	return a, w, nil
}

func (s *KeystoreStore) CreateWallet(name string) error {
	store, err := getStore(s.Path)
	if err != nil {
		return err
	}
	_, err = e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(store))
	if err != nil {
		return err
	}
	return nil
}

func (s *KeystoreStore) GetPK(w string, a string) ([]byte, error) {
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

func (s *KeystoreStore) SavePKToWallet(w string, a []byte, n string) error {
	wallet, err := getWallet(s.Path, w)
	if err != nil {
		return err
	}
	err = wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		return err
	}

	defer func() {
		err = wallet.(types.WalletLocker).Lock(context.Background())
	}()

	_, err = wallet.(types.WalletAccountImporter).ImportAccount(s.Ctx,
		n,
		a,
		//Always Use The First Password In Array
		s.Passphrases[0],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *KeystoreStore) GetPath() string {
	return s.Path
}

func (s *KeystoreStore) GetType() string {
	return s.Type
}

func newKeystoreStore(t string) (KeystoreStore, error) {
	s := KeystoreStore{}
	//Parse Wallet Type
	wt := viper.GetString(fmt.Sprintf("%s.wallet.type", t))
	utils.Log.Debug().Msgf("setting store type to %s", wt)
	s.Type = wt

	//Parse Store Path
	storePath := viper.GetString(fmt.Sprintf("%s.store.path", t))
	utils.Log.Debug().Msgf("setting store path to %s", storePath)
	if storePath == "" {
		return s, errors.New("keystore store path is empty")
	}
	s.Path = storePath

	//Parse Passphrases
	utils.Log.Debug().Msgf("getting passhphrases")
	passphrases, err := getAccountsPasswords(viper.GetString(fmt.Sprintf("%s.wallet.passphrases.path", t)))
	if err != nil {
		return s, err
	}
	utils.Log.Debug().Msgf("checking passhphrases len: %d", len(passphrases))
	if len(passphrases) == 0 {
		return s, errors.New("passhparases file for keystore store is empty")
	}

	// Cheking If Passphrases Index Is Set
	if viper.IsSet(fmt.Sprintf("%s.wallet.passphrases.index", t)) {
		index := viper.GetInt(fmt.Sprintf("%s.wallet.passphrases.index", t))
		passphrases = [][]byte{passphrases[index]}
	}

	s.Passphrases = passphrases

	return s, nil
}
