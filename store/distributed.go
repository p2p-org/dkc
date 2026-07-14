package store

import (
	"context"
	"fmt"

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

var _ AtomicStore = (*DistributedAtomicStore)(nil)

// newDistributedAtomicStore creates a distributed atomic store for a single peer
func newDistributedAtomicStore(
	ctx context.Context,
	side string,
	peerID uint64,
	peerName string,
	path string,
	walletType string,
	passphrases [][]byte,
	threshold uint32,
	peers map[uint64]string,
) (*DistributedAtomicStore, error) {
	s := &DistributedAtomicStore{
		Type:        walletType,
		Path:        path,
		Passphrases: passphrases,
		Ctx:         ctx,
		cache:       NewWalletCache(),
		PeerID:      peerID,
		PeerName:    peerName,
		Threshold:   threshold,
		Peers:       peers,
	}

	// Input stores serve reads: populate the cache up front
	if side == "input" {
		prefix := fmt.Sprintf("Peer %d", peerID)
		if err := s.cache.PopulateFromLocation(ctx, path, passphrases, prefix); err != nil {
			return nil, errors.Wrapf(err, "failed to populate distributed store cache for peer %d", peerID)
		}
	}

	return s, nil
}

func (s *DistributedAtomicStore) Create() error {
	// Nothing to do: the filesystem store creates directories lazily on write
	return nil
}

func (s *DistributedAtomicStore) GetAccounts() ([]AccountsData, []string, error) {
	return getWalletsAccountsMap(s.Ctx, s.Path)
}

func (s *DistributedAtomicStore) CreateWallet(name string) error {
	wallet, err := e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(newFSStore(s.Path)))
	if err != nil {
		return err
	}

	// Unlock wallet immediately after creation and keep it unlocked
	if err := wallet.(types.WalletLocker).Unlock(s.Ctx, nil); err != nil {
		// Don't return error, wallet might not require unlocking
		utils.Log.Warn().Err(err).Msgf("⚠️ Distributed Store [Peer %d]: failed to unlock wallet after creation: %s", s.PeerID, name)
	} else {
		utils.Log.Info().Msgf("🔓 Distributed Store [Peer %d]: wallet %s created and unlocked permanently", s.PeerID, name)
	}

	s.cache.PutWallet(wallet)

	return nil
}

func (s *DistributedAtomicStore) GetPrivateKey(walletName string, accountName string) ([]byte, error) {
	utils.Log.Info().Msgf("🔐 Distributed Store [Peer %d]: getting private key shard for account: %s/%s", s.PeerID, walletName, accountName)

	// Try to get account from cache first
	account, err := s.cache.FetchAccount(walletName, accountName)
	if err != nil {
		utils.Log.Warn().Msgf("⚠️ Distributed Store [Peer %d]: account not found in cache, falling back to direct access: %s/%s", s.PeerID, walletName, accountName)

		wallet, err := s.cache.GetOrOpenWallet(s.Path, walletName)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to get wallet: %s/%s", s.PeerID, walletName, accountName)
			return nil, err
		}
		if err := wallet.(types.WalletLocker).Unlock(s.Ctx, nil); err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to unlock wallet: %s/%s", s.PeerID, walletName, accountName)
			return nil, err
		}

		account, err = wallet.(types.WalletAccountByNameProvider).AccountByName(s.Ctx, accountName)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to get account: %s/%s", s.PeerID, walletName, accountName)
			return nil, err
		}
	}

	key, err := getAccountPrivateKey(s.Ctx, account, s.Passphrases)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to get private key: %s/%s", s.PeerID, walletName, accountName)
		return nil, err
	}

	utils.Log.Debug().Msgf("✅ Distributed Store [Peer %d]: successfully got key shard for account: %s/%s", s.PeerID, walletName, accountName)
	return key, nil
}

func (s *DistributedAtomicStore) SavePrivateKey(walletName string, accountName string, data any) error {
	utils.Log.Info().Msgf("💾 Distributed Store [Peer %d]: saving private key shard for account: %s/%s", s.PeerID, walletName, accountName)

	shardData, ok := data.(*DistributedAccountData)
	if !ok {
		return errors.Errorf("invalid data type for distributed store - expected *DistributedAccountData, got %T", data)
	}

	// Use the shared wallet instance so concurrent imports serialize on the
	// wallet's own mutex and index updates are never lost
	wallet, err := s.cache.GetOrOpenWallet(s.Path, walletName)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to get wallet for save: %s", s.PeerID, walletName)
		return err
	}

	// No need to unlock - wallet is already unlocked permanently after creation
	utils.Log.Debug().Msgf("🔧 Distributed Store [Peer %d]: importing distributed account with threshold %d", s.PeerID, shardData.Threshold)

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
		utils.Log.Error().Err(err).Msgf("❌ Distributed Store [Peer %d]: failed to import distributed account: %s/%s", s.PeerID, walletName, accountName)
		return err
	}

	utils.Log.Info().Msgf("✅ Distributed Store [Peer %d]: successfully saved key shard for account: %s/%s", s.PeerID, walletName, accountName)
	return nil
}

func (s *DistributedAtomicStore) GetPath() string {
	return s.Path
}

// GetThreshold returns the threshold for this distributed store
func (s *DistributedAtomicStore) GetThreshold() uint32 {
	return s.Threshold
}

// GetPeers returns all peers in the distributed store
func (s *DistributedAtomicStore) GetPeers() map[uint64]string {
	return s.Peers
}
