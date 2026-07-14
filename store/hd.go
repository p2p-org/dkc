package store

import (
	"context"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

// hdStore reads hierarchical-deterministic wallets. It is effectively
// input-only: HD keys are derived from a mnemonic and cannot be imported.
type hdStore struct {
	Type        string
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
	cache       *WalletCache
}

var _ AtomicStore = (*hdStore)(nil)

func newHDStore(ctx context.Context, side string) (*hdStore, error) {
	cfg, err := parseStoreConfig(side)
	if err != nil {
		return nil, err
	}
	s := &hdStore{
		Type:        cfg.WalletType,
		Path:        cfg.Path,
		Passphrases: cfg.Passphrases,
		Ctx:         ctx,
		cache:       NewWalletCache(),
	}

	// Input stores serve reads: populate the cache up front
	if side == "input" {
		if err := s.cache.PopulateFromLocation(ctx, s.Path, s.Passphrases, ""); err != nil {
			return nil, errors.Wrap(err, "failed to populate HD store cache")
		}
	}

	return s, nil
}

func (s *hdStore) Create() error {
	// Nothing to do: the filesystem store creates directories lazily on write
	return nil
}

func (s *hdStore) GetAccounts() ([]AccountsData, []string, error) {
	return getWalletsAccountsMap(s.Ctx, s.Path)
}

func (s *hdStore) CreateWallet(name string) error {
	_, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(newFSStore(s.Path)))
	return err
}

func (s *hdStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 HD Store: getting private key for account: %s/%s", walletName, accountName)

	account, err := s.cache.FetchAccount(walletName, accountName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ HD Store: failed to fetch account from cache: %s/%s", walletName, accountName)
		return nil, errors.Wrap(err, "account not found in cache")
	}

	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ HD Store: failed to get private key for account: %s/%s", walletName, accountName)
		return nil, err
	}

	utils.Log.Info().Msgf("✅ HD Store: successfully retrieved private key for account: %s/%s", walletName, accountName)
	return key, nil
}

func (s *hdStore) SavePrivateKey(walletName string, accountName string, data any) error {
	return errors.New("HD stores do not support saving private keys - keys are derived from mnemonic")
}

func (s *hdStore) GetPath() string {
	return s.Path
}
