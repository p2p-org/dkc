package store

import (
	"context"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// simpleStore is a single-path store whose wallets hold complete private
// keys. Non-deterministic and keystore wallet types share this behaviour
// and differ only in the wallet type passed to e2wallet.
type simpleStore struct {
	Type        string // wallet type from config, e.g. "non-deterministic" or "keystore"
	Label       string // log label, e.g. "ND Store"
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
	cache       *WalletCache
}

var _ AtomicStore = (*simpleStore)(nil)

func newSimpleStore(ctx context.Context, side string, label string) (*simpleStore, error) {
	cfg, err := parseStoreConfig(side)
	if err != nil {
		return nil, err
	}
	s := &simpleStore{
		Type:        cfg.WalletType,
		Label:       label,
		Path:        cfg.Path,
		Passphrases: cfg.Passphrases,
		Ctx:         ctx,
		cache:       NewWalletCache(),
	}

	// Input stores serve reads: populate the cache up front
	if side == "input" {
		if err := s.cache.PopulateFromLocation(ctx, s.Path, s.Passphrases, ""); err != nil {
			return nil, errors.Wrapf(err, "failed to populate %s cache", label)
		}
	}

	return s, nil
}

func (s *simpleStore) Create() error {
	// Nothing to do: the filesystem store creates directories lazily on write
	return nil
}

func (s *simpleStore) GetAccounts() ([]AccountsData, []string, error) {
	return getWalletsAccountsMap(s.Ctx, s.Path)
}

func (s *simpleStore) CreateWallet(name string) error {
	wallet, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(newFSStore(s.Path)))
	if err != nil {
		return err
	}

	// Unlock wallet immediately after creation and keep it unlocked
	if err := wallet.(types.WalletLocker).Unlock(s.Ctx, nil); err != nil {
		// Don't return error, wallet might not require unlocking
		utils.Log.Warn().Err(err).Msgf("⚠️ %s: failed to unlock wallet after creation: %s", s.Label, name)
	} else {
		utils.Log.Info().Msgf("🔓 %s: wallet %s created and unlocked permanently", s.Label, name)
	}

	s.cache.PutWallet(wallet)

	return nil
}

func (s *simpleStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 %s: getting private key for account: %s/%s", s.Label, walletName, accountName)

	account, err := s.cache.FetchAccount(walletName, accountName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ %s: failed to fetch account from cache: %s/%s", s.Label, walletName, accountName)
		return nil, errors.Wrap(err, "account not found in cache")
	}

	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ %s: failed to get private key for account: %s/%s", s.Label, walletName, accountName)
		return nil, err
	}

	utils.Log.Info().Msgf("✅ %s: successfully retrieved private key for account: %s/%s", s.Label, walletName, accountName)
	return key, nil
}

func (s *simpleStore) SavePrivateKey(walletName string, accountName string, data any) error {
	privateKey, ok := data.([]byte)
	if !ok {
		return errors.Errorf("invalid data type for %s - expected []byte", s.Label)
	}

	utils.Log.Debug().Msgf("💾 %s: importing account: %s/%s", s.Label, walletName, accountName)

	// Use the shared wallet instance so concurrent imports serialize on the
	// wallet's own mutex and index updates are never lost
	wallet, err := s.cache.GetOrOpenWallet(s.Path, walletName)
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

	return err
}

func (s *simpleStore) GetPath() string {
	return s.Path
}
