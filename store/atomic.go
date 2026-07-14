package store

import (
	"github.com/p2p-org/dkc/crypto/bls"
	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
)

// DistributedAccountData holds the data needed for ImportDistributedAccount
type DistributedAccountData struct {
	ParticipantShard []byte
	Threshold        uint32
	MasterPKs        [][]byte
	Peers            map[uint64]string
}

// AtomicStore represents a single physical store (one path, one cache).
// For distributed stores, each peer is a separate AtomicStore.
type AtomicStore interface {
	// Input operations
	GetPrivateKey(walletName, accountName string) ([]byte, error) // Returns full key OR shard
	GetAccounts() ([]AccountsData, []string, error)

	// Output operations
	Create() error
	CreateWallet(name string) error
	SavePrivateKey(walletName, accountName string, data any) error // Saves full key ([]byte) OR distributed data (*DistributedAccountData)

	// Metadata
	GetPath() string
}

// ComposedStore wraps one or more AtomicStores
// - Single AtomicStore: HD/ND/Keystore
// - Multiple AtomicStores: Distributed (one per peer)
type ComposedStore struct {
	atomicStores []AtomicStore
	storeType    string // for logging/debugging
}

// NewComposedStore creates a new composed store
func NewComposedStore(atomicStores []AtomicStore, storeType string) *ComposedStore {
	return &ComposedStore{
		atomicStores: atomicStores,
		storeType:    storeType,
	}
}

// GetPrivateKey retrieves a private key, handling single/distributed logic
func (cs *ComposedStore) GetPrivateKey(walletName, accountName string) ([]byte, error) {
	// Single store: HD/ND/Keystore - return full key
	if len(cs.atomicStores) == 1 {
		return cs.atomicStores[0].GetPrivateKey(walletName, accountName)
	}

	// Multiple stores: Distributed - collect shards and combine
	shards := make(map[uint64][]byte)
	for _, store := range cs.atomicStores {
		distributedStore, ok := store.(*DistributedAtomicStore)
		if !ok {
			return nil, errors.New("expected DistributedAtomicStore for multi-store composition")
		}

		shard, err := store.GetPrivateKey(walletName, accountName)
		if err != nil {
			return nil, err
		}

		shards[distributedStore.PeerID] = shard
	}

	utils.Log.Info().Msgf("🧩 ComposedStore: combining %d key shards for account: %s/%s", len(shards), walletName, accountName)

	combinedKey, err := bls.Combine(shards)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ ComposedStore: failed to combine key shards for account: %s/%s", walletName, accountName)
		return nil, err
	}

	utils.Log.Info().Msgf("✅ ComposedStore: successfully combined private key for account: %s/%s", walletName, accountName)
	return combinedKey, nil
}

// SavePrivateKey saves a private key, handling single/distributed logic
func (cs *ComposedStore) SavePrivateKey(walletName, accountName string, privateKey []byte) error {
	// Single store: save full key as-is
	if len(cs.atomicStores) == 1 {
		return cs.atomicStores[0].SavePrivateKey(walletName, accountName, privateKey)
	}

	// Multiple stores: split key into shards and save to each peer
	first, ok := cs.atomicStores[0].(*DistributedAtomicStore)
	if !ok {
		return errors.New("expected DistributedAtomicStore for multi-store composition")
	}
	threshold := first.GetThreshold()
	peers := first.GetPeers()

	utils.Log.Info().Msgf("🔪 ComposedStore: splitting private key for distributed store: %s/%s", walletName, accountName)
	utils.Log.Debug().Msgf("📊 ComposedStore: using threshold %d with %d peers", threshold, len(peers))

	masterSKs, masterPKs, err := bls.Split(privateKey, threshold)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ ComposedStore: failed to split private key for: %s/%s", walletName, accountName)
		return err
	}

	peerIDs := make([]uint64, 0, len(cs.atomicStores))
	for _, store := range cs.atomicStores {
		peerIDs = append(peerIDs, store.(*DistributedAtomicStore).PeerID)
	}

	// Derive each participant's shard from the master keys
	participants, err := bls.SetupParticipants(masterSKs, peerIDs)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ ComposedStore: failed to setup participants for: %s/%s", walletName, accountName)
		return err
	}

	for _, store := range cs.atomicStores {
		distributedStore := store.(*DistributedAtomicStore)

		utils.Log.Debug().Msgf("💾 ComposedStore: saving shard to peer %d for account: %s/%s", distributedStore.PeerID, walletName, accountName)

		shardData := &DistributedAccountData{
			ParticipantShard: participants[distributedStore.PeerID],
			Threshold:        threshold,
			MasterPKs:        masterPKs,
			Peers:            peers,
		}

		if err := store.SavePrivateKey(walletName, accountName, shardData); err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ComposedStore: failed to save shard to peer %d for: %s/%s", distributedStore.PeerID, walletName, accountName)
			return err
		}
	}

	utils.Log.Info().Msgf("✅ ComposedStore: successfully saved distributed private key for account: %s/%s", walletName, accountName)
	return nil
}

// GetAccounts returns all accounts from the first atomic store
// (for distributed stores, all peers have the same account structure)
func (cs *ComposedStore) GetAccounts() ([]AccountsData, []string, error) {
	return cs.atomicStores[0].GetAccounts()
}

// Create creates all underlying atomic stores
func (cs *ComposedStore) Create() error {
	for _, store := range cs.atomicStores {
		if err := store.Create(); err != nil {
			return err
		}
	}
	return nil
}

// CreateWallet creates a wallet in all underlying atomic stores
func (cs *ComposedStore) CreateWallet(name string) error {
	for _, store := range cs.atomicStores {
		if err := store.CreateWallet(name); err != nil {
			return err
		}
	}
	return nil
}

// GetType returns the store type for logging
func (cs *ComposedStore) GetType() string {
	return cs.storeType
}

// GetPath returns the primary path (first atomic store)
func (cs *ComposedStore) GetPath() string {
	if len(cs.atomicStores) > 0 {
		return cs.atomicStores[0].GetPath()
	}
	return ""
}
