package store

import (
	"context"

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

// AtomicStore represents a single physical store (one path, one cache)
// For distributed stores, each peer is a separate AtomicStore
type AtomicStore interface {
	// Input operations
	GetPrivateKey(walletName, accountName string) ([]byte, error) // Returns full key OR shard
	GetAccounts() ([]AccountsData, []string, error)
	GetCache() *WalletCache

	// Output operations
	Create() error
	CreateWallet(name string) error
	SavePrivateKey(walletName, accountName string, data interface{}) error // Saves full key ([]byte) OR distributed data (*DistributedAccountData)

	// Metadata
	GetType() string
	GetPath() string
	GetContext() context.Context
	SetContext(ctx context.Context)
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
	if len(cs.atomicStores) == 1 {
		// Single store: HD/ND/Keystore - return full key
		return cs.atomicStores[0].GetPrivateKey(walletName, accountName)
	} else {
		// Multiple stores: Distributed - collect shards and combine
		// For distributed stores, we need to map peer IDs correctly
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

			// Use the actual peer ID from the distributed store, not the array index
			shards[distributedStore.PeerID] = shard
		}

		// Add logging for debugging
		utils.Log.Info().Msgf("🧩 ComposedStore: Combining %d key shards for account: %s/%s", len(shards), walletName, accountName)
		for peerID := range shards {
			utils.Log.Debug().Msgf("🔑 ComposedStore: Got shard from peer %d", peerID)
		}

		// Use existing BLS combine function with proper peer IDs
		combinedKey, err := bls.Combine(cs.atomicStores[0].GetContext(), shards)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ComposedStore: Failed to combine key shards for account: %s/%s", walletName, accountName)
			return nil, err
		}

		utils.Log.Info().Msgf("✅ ComposedStore: Successfully combined private key for account: %s/%s", walletName, accountName)
		return combinedKey, nil
	}
}

// SavePrivateKey saves a private key, handling single/distributed logic
func (cs *ComposedStore) SavePrivateKey(walletName, accountName string, privateKey []byte) error {
	if len(cs.atomicStores) == 1 {
		// Single store: save full key as-is (pass as []byte)
		return cs.atomicStores[0].SavePrivateKey(walletName, accountName, privateKey)
	} else {
		// Multiple stores: split key into shards and save to each peer
		ctx := cs.atomicStores[0].GetContext()

		// Get distributed store configuration
		threshold := cs.getDistributedThreshold()
		peers := cs.getDistributedPeers()

		utils.Log.Info().Msgf("🔪 ComposedStore: Splitting private key for distributed store: %s/%s", walletName, accountName)
		utils.Log.Debug().Msgf("📊 ComposedStore: Using threshold %d with %d peers", threshold, len(peers))

		// Split the key into master secrets
		masterSKs, masterPKs, err := bls.Split(ctx, privateKey, threshold)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ComposedStore: Failed to split private key for: %s/%s", walletName, accountName)
			return err
		}

		// Extract peer IDs from distributed stores
		var peerIDs []uint64
		for _, store := range cs.atomicStores {
			distributedStore := store.(*DistributedAtomicStore)
			peerIDs = append(peerIDs, distributedStore.PeerID)
		}

		// Setup participants (create shards for each peer)
		participants, err := bls.SetupParticipants(masterSKs, masterPKs, peerIDs, len(cs.atomicStores))
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ComposedStore: Failed to setup participants for: %s/%s", walletName, accountName)
			return err
		}

		// Save each participant's shard to corresponding peer
		for _, store := range cs.atomicStores {
			distributedStore := store.(*DistributedAtomicStore)
			participantShard := participants[distributedStore.PeerID]

			utils.Log.Debug().Msgf("💾 ComposedStore: Saving shard to peer %d for account: %s/%s", distributedStore.PeerID, walletName, accountName)

			// Create the data structure expected by ImportDistributedAccount
			shardData := &DistributedAccountData{
				ParticipantShard: participantShard,
				Threshold:        threshold,
				MasterPKs:        masterPKs,
				Peers:            peers,
			}

			if err := store.SavePrivateKey(walletName, accountName, shardData); err != nil {
				utils.Log.Error().Err(err).Msgf("❌ ComposedStore: Failed to save shard to peer %d for: %s/%s", distributedStore.PeerID, walletName, accountName)
				return err
			}
		}

		utils.Log.Info().Msgf("✅ ComposedStore: Successfully saved distributed private key for account: %s/%s", walletName, accountName)
		return nil
	}
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

// Helper methods for distributed store operations
func (cs *ComposedStore) getDistributedThreshold() uint32 {
	if len(cs.atomicStores) > 0 {
		if distributedStore, ok := cs.atomicStores[0].(*DistributedAtomicStore); ok {
			return distributedStore.GetThreshold()
		}
	}
	return 2 // fallback
}

func (cs *ComposedStore) getDistributedPeers() map[uint64]string {
	if len(cs.atomicStores) > 0 {
		if distributedStore, ok := cs.atomicStores[0].(*DistributedAtomicStore); ok {
			return distributedStore.GetPeers()
		}
	}
	return make(map[uint64]string) // fallback
}
