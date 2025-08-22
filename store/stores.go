package store

import (
	"context"
	"fmt"
	"regexp"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Types moved from distributed_old.go
type Peers map[uint64]Peer
type Peer struct {
	Name        string
	Passphrases struct {
		Path  string
		Index int
	}
}

type Threshold uint32

// InputStoreComposed creates a new ComposedStore for input operations
func InputStoreComposed(ctx context.Context, storeType string) (*ComposedStore, error) {
	switch storeType {
	case "hierarchical deterministic":
		atomic, err := createHDAtomicStore(ctx, "input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create HD atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "HD"), nil

	case "non-deterministic":
		atomic, err := createNDAtomicStore(ctx, "input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ND atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "ND"), nil

	case "keystore":
		atomic, err := createKeystoreAtomicStore(ctx, "input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Keystore atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "Keystore"), nil

	case "distributed":
		atomics, err := createDistributedAtomicStores(ctx, "input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Distributed atomic stores")
		}
		return NewComposedStore(atomics, "Distributed"), nil

	default:
		return nil, errors.New("incorrect input wallet type")
	}
}

// OutputStoreComposed creates a new ComposedStore for output operations
func OutputStoreComposed(ctx context.Context, storeType string) (*ComposedStore, error) {
	switch storeType {
	case "hierarchical deterministic":
		atomic, err := createHDAtomicStore(ctx, "output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create HD atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "HD"), nil

	case "non-deterministic":
		atomic, err := createNDAtomicStore(ctx, "output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ND atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "ND"), nil

	case "keystore":
		atomic, err := createKeystoreAtomicStore(ctx, "output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Keystore atomic store")
		}
		return NewComposedStore([]AtomicStore{atomic}, "Keystore"), nil

	case "distributed":
		atomics, err := createDistributedAtomicStores(ctx, "output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Distributed atomic stores")
		}
		return NewComposedStore(atomics, "Distributed"), nil

	default:
		return nil, errors.New("incorrect output wallet type")
	}
}

// Helper functions to create atomic stores

func createHDAtomicStore(ctx context.Context, storeType string) (AtomicStore, error) {
	store, err := newHDStore(storeType)
	if err != nil {
		return nil, err
	}
	store.SetContext(ctx)

	// Populate cache for input stores
	if storeType == "input" {
		utils.Log.Info().Msgf("💾 HD Store: Starting cache population for input store")
		err = store.GetCache().PopulateFromLocation(ctx, store.GetPath(), store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ HD Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate HD store cache")
		}
		utils.Log.Info().Msgf("✅ HD Store: Cache population completed for input store")
	}

	return &store, nil
}

func createNDAtomicStore(ctx context.Context, storeType string) (AtomicStore, error) {
	store, err := newNDStore(storeType)
	if err != nil {
		return nil, err
	}
	store.SetContext(ctx)

	// Populate cache for input stores
	if storeType == "input" {
		utils.Log.Info().Msgf("💾 ND Store: Starting cache population for input store")
		err = store.GetCache().PopulateFromLocation(ctx, store.GetPath(), store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ND Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate ND store cache")
		}
		utils.Log.Info().Msgf("✅ ND Store: Cache population completed for input store")
	}

	return &store, nil
}

func createKeystoreAtomicStore(ctx context.Context, storeType string) (AtomicStore, error) {
	store, err := newKeystoreStore(storeType)
	if err != nil {
		return nil, err
	}
	store.SetContext(ctx)

	// Populate cache for input stores
	if storeType == "input" {
		utils.Log.Info().Msgf("💾 Keystore Store: Starting cache population for input store")
		err = store.GetCache().PopulateFromLocation(ctx, store.GetPath(), store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Keystore Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate keystore store cache")
		}
		utils.Log.Info().Msgf("✅ Keystore Store: Cache population completed for input store")
	}

	return &store, nil
}

func createDistributedAtomicStores(ctx context.Context, storeType string) ([]AtomicStore, error) {
	// Parse configuration similar to existing newDistributedStore
	walletType := viper.GetString(fmt.Sprintf("%s.wallet.type", storeType))
	storePath := viper.GetString(fmt.Sprintf("%s.store.path", storeType))
	if storePath == "" {
		return nil, errors.New("distributed store path is empty")
	}

	// Parse Peers
	var peers Peers
	err := viper.UnmarshalKey(fmt.Sprintf("%s.wallet.peers", storeType), &peers)
	if err != nil {
		return nil, err
	}
	if len(peers) < 2 {
		return nil, errors.New("number of peers for distributed store is less than 2")
	}

	// Parse Threshold
	var threshold Threshold
	err = viper.UnmarshalKey(fmt.Sprintf("%s.wallet.threshold", storeType), &threshold)
	if err != nil {
		return nil, err
	}
	if uint32(threshold) <= uint32(len(peers)/2) || uint32(threshold) > uint32(len(peers)) {
		return nil, errors.New("invalid threshold value for distributed store")
	}

	// Create atomic stores for each peer
	var atomicStores []AtomicStore
	peersMap := make(map[uint64]string)
	res, err := regexp.Compile(`:.*`)
	if err != nil {
		return nil, err
	}

	for id, peer := range peers {
		// Parse passphrases for this peer
		passphrases, err := getAccountsPasswords(peer.Passphrases.Path)
		if err != nil {
			return nil, err
		}
		if len(passphrases) == 0 {
			return nil, errors.New("passphrases file for distributed peer is empty")
		}

		// Check if passphrases index is set
		if viper.IsSet(fmt.Sprintf("%s.peers.%d.passphrases.index", storeType, id)) {
			index := viper.GetInt(fmt.Sprintf("%s.peers.%d.passphrases.index", storeType, id))
			passphrases = [][]byte{passphrases[index]}
		}

		peersMap[id] = peer.Name
		peerPath := storePath + "/" + res.ReplaceAllString(peer.Name, "")

		// Create atomic store for this peer
		atomicStore := NewDistributedAtomicStore(
			id,
			peer.Name,
			peerPath,
			walletType,
			passphrases,
			uint32(threshold),
			peersMap,
		)
		atomicStore.SetContext(ctx)

		// Populate cache for input stores
		if storeType == "input" {
			utils.Log.Info().Msgf("🔄 Distributed Store: Populating cache for peer %d (%s)", id, peer.Name)
			peerPrefix := fmt.Sprintf("Peer %d", id)
			err = atomicStore.GetCache().PopulateFromLocationWithPrefix(ctx, peerPath, passphrases, peerPrefix)
			if err != nil {
				utils.Log.Error().Err(err).Msgf("❌ Distributed Store: Failed to populate cache for peer %d", id)
				return nil, errors.Wrap(err, fmt.Sprintf("failed to populate distributed store cache for peer %d", id))
			}
			utils.Log.Info().Msgf("✅ Distributed Store: Cache population completed for peer %d (%s)", id, peer.Name)
		}

		atomicStores = append(atomicStores, atomicStore)
	}

	if storeType == "input" {
		utils.Log.Info().Msgf("🎉 Distributed Store: All peer caches populated successfully")
	}

	return atomicStores, nil
}
