package store

import (
	"context"
	"fmt"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
)

type InputStore interface {
	// Get Wallets Names And Account Names
	GetWalletsAccountsMap() ([]AccountsData, []string, error)
	// Get Private Key From Wallet Using Account Name
	GetPrivateKey(walletName string, accountName string) ([]byte, error)
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
	// Get Wallet Cache
	GetWalletCache() *WalletCache
}

type OutputStore interface {
	// Create Store
	Create() error
	// Create New Wallet
	CreateWallet(name string) error
	// Save Private Key To Wallet
	SavePrivateKey(walletName string, accountName string, privateKey []byte) error
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
}

func InputStoreInit(ctx context.Context, storeType string) (InputStore, error) {
	switch storeType {
	case "distributed":
		store, err := newDistributedStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as input store")
		}
		store.Ctx = ctx

		// Populate cache for each peer (each peer has different private key shards)
		utils.Log.Info().Msgf("💾 Distributed Store: Starting cache population for all %d peers", len(store.Peers))
		for id := range store.Peers {
			utils.Log.Info().Msgf("🔄 Distributed Store: Populating cache for peer %d (%s)", id, store.Peers[id])
			peerCache := store.peerCaches[id]
			if peerCache == nil {
				utils.Log.Error().Msgf("❌ Distributed Store: No cache found for peer %d", id)
				return nil, errors.New("peer cache not initialized")
			}

			peerPrefix := fmt.Sprintf("Peer %d", id)
			err = peerCache.PopulateFromLocationWithPrefix(ctx, store.PeersPaths[id], store.PeersPasswords[id], peerPrefix)
			if err != nil {
				utils.Log.Error().Err(err).Msgf("❌ Distributed Store: Failed to populate cache for peer %d", id)
				return nil, errors.Wrap(err, fmt.Sprintf("failed to populate distributed store cache for peer %d", id))
			}
			utils.Log.Info().Msgf("✅ Distributed Store: Cache population completed for peer %d (%s)", id, store.Peers[id])
		}
		utils.Log.Info().Msgf("🎉 Distributed Store: All peer caches populated successfully")

		return &store, nil
	case "hierarchical deterministic":
		store, err := newHDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init hierarchial deterministic store as input store")
		}
		store.Ctx = ctx

		// Populate cache
		utils.Log.Info().Msgf("💾 HD Store: Starting cache population for input store")
		err = store.cache.PopulateFromLocation(ctx, store.Path, store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ HD Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate HD store cache")
		}
		utils.Log.Info().Msgf("✅ HD Store: Cache population completed for input store")

		return &store, nil
	case "non-deterministic":
		store, err := newNDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as input store")
		}
		store.Ctx = ctx

		// Populate cache
		utils.Log.Info().Msgf("💾 ND Store: Starting cache population for input store")
		err = store.cache.PopulateFromLocation(ctx, store.Path, store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ ND Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate ND store cache")
		}
		utils.Log.Info().Msgf("✅ ND Store: Cache population completed for input store")

		return &store, nil
	case "keystore":
		store, err := newKeystoreStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init keystore store as input store")
		}
		store.Ctx = ctx

		// Populate cache
		utils.Log.Info().Msgf("💾 Keystore Store: Starting cache population for input store")
		err = store.cache.PopulateFromLocation(ctx, store.Path, store.Passphrases)
		if err != nil {
			utils.Log.Error().Err(err).Msgf("❌ Keystore Store: Failed to populate cache")
			return nil, errors.Wrap(err, "failed to populate keystore store cache")
		}
		utils.Log.Info().Msgf("✅ Keystore Store: Cache population completed for input store")

		return &store, nil
	default:
		return nil, errors.New("incorrect input wallet type")
	}

}

func OutputStoreInit(ctx context.Context, storeType string) (OutputStore, error) {
	switch storeType {
	case "distributed":
		store, err := newDistributedStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	case "non-deterministic":
		store, err := newNDStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	case "keystore":
		store, err := newKeystoreStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init keystore store as output store")
		}
		store.Ctx = ctx
		return &store, nil
	default:
		return nil, errors.New("incorrect output wallet type")
	}

}
