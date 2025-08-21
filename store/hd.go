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
	cache       *WalletCache
}

func (s *HDStore) Create() error {
	_, err := createStore(s.Path)
	if err != nil {
		return err
	}

	return nil
}
func (s *HDStore) GetWalletsAccountsMap() ([]AccountsData, []string, error) {
	account, wallet, err := getWalletsAccountsMap(s.Ctx, s.Path)
	if err != nil {
		return nil, nil, err
	}

	return account, wallet, nil
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

func (s *HDStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 HD Store: Getting private key for account: %s/%s", walletName, accountName)

	// Try to get account from cache first
	account, err := s.cache.FetchAccount(walletName, accountName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ HD Store: Failed to fetch account from cache: %s/%s", walletName, accountName)
		return nil, errors.Wrap(err, "account not found in cache")
	}

	utils.Log.Debug().Msgf("🔓 HD Store: Extracting private key for account: %s/%s", walletName, accountName)
	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ HD Store: Failed to get private key for account: %s/%s", walletName, accountName)
		return nil, err
	}

	utils.Log.Info().Msgf("✅ HD Store: Successfully retrieved private key for account: %s/%s", walletName, accountName)
	return key, nil
}

func (s *HDStore) GetPath() string {
	return s.Path
}

func (s *HDStore) GetType() string {
	return s.Type
}

func (s *HDStore) GetWalletCache() *WalletCache {
	return s.cache
}

func newHDStore(storeType string) (HDStore, error) {
	store := HDStore{
		cache: NewWalletCache(),
	}
	//Parse Wallet Type
	walletType := viper.GetString(fmt.Sprintf("%s.wallet.type", storeType))
	utils.Log.Debug().Msgf("setting store type to %s", walletType)
	store.Type = walletType

	//Parse Store Path
	storePath := viper.GetString(fmt.Sprintf("%s.store.path", storeType))
	utils.Log.Debug().Msgf("setting store path to %s", storePath)
	if storePath == "" {
		return store, errors.New("hd store path is empty")
	}
	store.Path = storePath

	//Parse Passphrases
	utils.Log.Debug().Msgf("getting passhphrases")
	passphrases, err := getAccountsPasswords(
		viper.GetString(fmt.Sprintf("%s.wallet.passphrases.path", storeType)),
	)
	if err != nil {
		return store, err
	}
	utils.Log.Debug().Msgf("checking passhphrases len: %d", len(passphrases))
	if len(passphrases) == 0 {
		return store, errors.New("passhparases file for hd store is empty")
	}

	// Cheking If Passphrases Index Is Set
	if viper.IsSet(fmt.Sprintf("%s.wallet.passphrases.index", storeType)) {
		index := viper.GetInt(fmt.Sprintf("%s.wallet.passphrases.index", storeType))
		passphrases = [][]byte{passphrases[index]}
	}

	store.Passphrases = passphrases

	return store, nil
}
