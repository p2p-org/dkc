package store

import (
	"context"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// DistributedAtomicStore represents a single peer in a distributed store
type DistributedAtomicStore struct {
	Type        string
	Path        string
	Passphrases [][]byte
	Ctx         context.Context
	cache       *WalletCache

	// Distributed store specific fields
	PeerID    uint64
	PeerName  string
	Threshold uint32
	Peers     map[uint64]string // All peers in the distributed store
}

// Ensure DistributedAtomicStore implements AtomicStore interface
var _ AtomicStore = (*DistributedAtomicStore)(nil)

func (s *DistributedAtomicStore) Create() error {
	_, err := createStore(s.Path)
	if err != nil {
		return err
	}
	return nil
}

func (s *DistributedAtomicStore) GetAccounts() ([]AccountsData, []string, error) {
	account, wallet, err := getWalletsAccountsMap(s.Ctx, s.Path)
	if err != nil {
		return nil, nil, err
	}
	return account, wallet, nil
}

func (s *DistributedAtomicStore) CreateWallet(name string) error {
	store, err := getStore(s.Path)
	if err != nil {
		return err
	}
	wallet, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(store))
	if err != nil {
		return err
	}

	// Unlock wallet immediately after creation and keep it unlocked
	err = wallet.(types.WalletLocker).Unlock(s.Ctx, nil)
	if err != nil {
		utils.Log.Warn().Err(err).Msgf("⚠️ Distributed Atomic Store [Peer %d]: Failed to unlock wallet after creation: %s", s.PeerID, name)
		// Don't return error, wallet might not require unlocking
	} else {
		utils.Log.Info().Msgf("🔓 Distributed Atomic Store [Peer %d]: Wallet %s created and unlocked permanently", s.PeerID, name)
	}

	return nil
}

func (s *DistributedAtomicStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 Distributed Atomic Store [Peer %d]: Getting private key shard for account: %s/%s", s.PeerID, walletName, accountName)

	// Try to get account from cache first
	account, err := s.cache.FetchAccount(walletName, accountName)
	if err == nil {
		utils.Log.Debug().Msgf("💾 Distributed Atomic Store [Peer %d]: Found account in cache: %s/%s", s.PeerID, walletName, accountName)

		// Extract private key shard from cached account
		utils.Log.Debug().Msgf("🔓 Distributed Atomic Store [Peer %d]: Extracting key shard from cached account: %s/%s", s.PeerID, walletName, accountName)
		key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to get private key from cached account: %s/%s", s.PeerID, walletName, accountName)
			return nil, err
		}

		utils.Log.Debug().Msgf("✅ Distributed Atomic Store [Peer %d]: Successfully got key shard from cache for account: %s/%s", s.PeerID, walletName, accountName)
		return key, nil
	} else {
		utils.Log.Warn().Msgf("⚠️ Distributed Atomic Store [Peer %d]: Account not found in cache, falling back to direct access: %s/%s", s.PeerID, walletName, accountName)
	}

	// Fallback to direct wallet access if cache miss
	wallet, err := getWallet(s.Path, walletName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to get wallet: %s/%s", s.PeerID, walletName, accountName)
		return nil, err
	}
	err = wallet.(types.WalletLocker).Unlock(s.Ctx, nil)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to unlock wallet: %s/%s", s.PeerID, walletName, accountName)
		return nil, err
	}

	account, err = wallet.(types.WalletAccountByNameProvider).AccountByName(s.Ctx, accountName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to get account: %s/%s", s.PeerID, walletName, accountName)
		return nil, err
	}

	utils.Log.Debug().Msgf("🔓 Distributed Atomic Store [Peer %d]: Extracting key shard from direct wallet access: %s/%s", s.PeerID, walletName, accountName)
	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to get private key: %s/%s", s.PeerID, walletName, accountName)
		return nil, err
	}

	utils.Log.Debug().Msgf("✅ Distributed Atomic Store [Peer %d]: Successfully got key shard from direct access for account: %s/%s", s.PeerID, walletName, accountName)
	return key, nil
}

func (s *DistributedAtomicStore) SavePrivateKey(walletName string, accountName string, data interface{}) error {
	utils.Log.Info().Msgf("💾 Distributed Atomic Store [Peer %d]: Saving private key shard for account: %s/%s", s.PeerID, walletName, accountName)

	// Extract distributed account data
	shardData, ok := data.(*DistributedAccountData)
	if !ok {
		utils.Log.Error().Msgf("❌ Distributed Atomic Store [Peer %d]: Expected *DistributedAccountData, got %T", s.PeerID, data)
		return errors.New("invalid data type for distributed store - expected *DistributedAccountData")
	}

	// Prevent concurrent file corruption with wallet-level mutex
	walletPath := s.Path + "/" + walletName
	fileMu := getFileMutex(walletPath)
	fileMu.Lock()
	defer fileMu.Unlock()

	wallet, err := getWallet(s.Path, walletName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to get wallet for save: %s", s.PeerID, walletName)
		return err
	}

	// No need to unlock - wallet is already unlocked permanently after creation
	utils.Log.Debug().Msgf("🔧 Distributed Atomic Store [Peer %d]: Importing distributed account with threshold %d (wallet already unlocked)", s.PeerID, shardData.Threshold)

	_, err = wallet.(types.WalletDistributedAccountImporter).ImportDistributedAccount(
		s.Ctx,
		accountName,
		shardData.ParticipantShard,
		shardData.Threshold,
		shardData.MasterPKs,
		shardData.Peers,
		s.Passphrases[0], // Always use the first password
	)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Atomic Store [Peer %d]: Failed to import distributed account: %s/%s", s.PeerID, walletName, accountName)
		return err
	}

	utils.Log.Info().Msgf("✅ Distributed Atomic Store [Peer %d]: Successfully saved key shard for account: %s/%s", s.PeerID, walletName, accountName)
	return nil
}

func (s *DistributedAtomicStore) GetType() string {
	return s.Type
}

func (s *DistributedAtomicStore) GetPath() string {
	return s.Path
}

func (s *DistributedAtomicStore) GetCache() *WalletCache {
	return s.cache
}

func (s *DistributedAtomicStore) GetContext() context.Context {
	return s.Ctx
}

func (s *DistributedAtomicStore) SetContext(ctx context.Context) {
	s.Ctx = ctx
}

// GetThreshold returns the threshold for this distributed store
func (s *DistributedAtomicStore) GetThreshold() uint32 {
	return s.Threshold
}

// GetPeers returns all peers in the distributed store
func (s *DistributedAtomicStore) GetPeers() map[uint64]string {
	return s.Peers
}

// NewDistributedAtomicStore creates a new distributed atomic store for a single peer
func NewDistributedAtomicStore(peerID uint64, peerName string, path string, storeType string, passphrases [][]byte, threshold uint32, peers map[uint64]string) *DistributedAtomicStore {
	return &DistributedAtomicStore{
		Type:        storeType,
		Path:        path,
		Passphrases: passphrases,
		cache:       NewWalletCache(),
		PeerID:      peerID,
		PeerName:    peerName,
		Threshold:   threshold,
		Peers:       peers,
	}
}
