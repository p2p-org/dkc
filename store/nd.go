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

type NDStore struct {
	Type        string
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
	cache       *WalletCache
}

// Ensure NDStore implements AtomicStore interface
var _ AtomicStore = (*NDStore)(nil)

func (s *NDStore) Create() error {
	_, err := createStore(s.Path)
	if err != nil {
		return err
	}

	return nil
}

// GetAccounts implements AtomicStore interface
func (s *NDStore) GetAccounts() ([]AccountsData, []string, error) {
	account, wallet, err := getWalletsAccountsMap(s.Ctx, s.Path)
	if err != nil {
		return nil, nil, err
	}
	return account, wallet, nil
}

func (s *NDStore) CreateWallet(name string) error {
	store, err := getStore(s.Path)
	if err != nil {
		return err
	}
	wallet, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(store))
	if err != nil {
		return err
	}

	// Unlock wallet immediately after creation and keep it unlocked
	err = wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		utils.Log.Warn().Err(err).Msgf("⚠️ ND Store: Failed to unlock wallet after creation: %s", name)
		// Don't return error, wallet might not require unlocking
	} else {
		utils.Log.Info().Msgf("🔓 ND Store: Wallet %s created and unlocked permanently", name)
	}

	return nil
}

func (s *NDStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 ND Store: Getting private key for account: %s/%s", walletName, accountName)

	// Try to get account from cache first
	account, err := s.cache.FetchAccount(walletName, accountName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ ND Store: Failed to fetch account from cache: %s/%s", walletName, accountName)
		return nil, errors.Wrap(err, "account not found in cache")
	}

	utils.Log.Debug().Msgf("🔓 ND Store: Extracting private key for account: %s/%s", walletName, accountName)
	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ ND Store: Failed to get private key for account: %s/%s", walletName, accountName)
		return nil, err
	}

	utils.Log.Info().Msgf("✅ ND Store: Successfully retrieved private key for account: %s/%s", walletName, accountName)
	return key, nil
}

func (s *NDStore) SavePrivateKey(walletName string, accountName string, data interface{}) error {
	// Extract private key from interface
	privateKey, ok := data.([]byte)
	if !ok {
		return errors.New("invalid data type for ND store - expected []byte")
	}

	// Prevent concurrent file corruption with wallet-level mutex
	walletPath := s.Path + "/" + walletName
	fileMu := getFileMutex(walletPath)
	fileMu.Lock()
	defer fileMu.Unlock()

	utils.Log.Debug().Msgf("💾 ND Store: Importing account (wallet already unlocked): %s/%s", walletName, accountName)

	wallet, err := getWallet(s.Path, walletName)
	if err != nil {
		return err
	}

	// No need to unlock - wallet is already unlocked permanently after creation

	_, err = wallet.(types.WalletAccountImporter).ImportAccount(s.Ctx,
		accountName,
		privateKey,
		//Always Use The First Password In Array
		s.Passphrases[0],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *NDStore) GetPath() string {
	return s.Path
}

func (s *NDStore) GetType() string {
	return s.Type
}

// GetCache implements AtomicStore interface
func (s *NDStore) GetCache() *WalletCache {
	return s.cache
}

// GetContext implements AtomicStore interface
func (s *NDStore) GetContext() context.Context {
	return s.Ctx
}

// SetContext implements AtomicStore interface
func (s *NDStore) SetContext(ctx context.Context) {
	s.Ctx = ctx
}

func newNDStore(storeType string) (NDStore, error) {
	store := NDStore{
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
		return store, errors.New("nd store path is empty")
	}
	store.Path = storePath

	//Parse Passphrases
	utils.Log.Debug().Msgf("getting passhphrases")
	passphrases, err := getAccountsPasswords(viper.GetString(fmt.Sprintf("%s.wallet.passphrases.path", storeType)))
	if err != nil {
		return store, err
	}
	utils.Log.Debug().Msgf("checking passhphrases len: %d", len(passphrases))
	if len(passphrases) == 0 {
		return store, errors.New("passhparases file for nd store is empty")
	}

	// Cheking If Passphrases Index Is Set
	if viper.IsSet(fmt.Sprintf("%s.wallet.passphrases.index", storeType)) {
		index := viper.GetInt(fmt.Sprintf("%s.wallet.passphrases.index", storeType))
		passphrases = [][]byte{passphrases[index]}
	}

	store.Passphrases = passphrases

	return store, nil
}
