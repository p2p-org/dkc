package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Peer is a single peer entry of a distributed store config
type Peer struct {
	Name        string
	Passphrases struct {
		Path  string
		Index int
	}
}

// InputStoreComposed creates a new ComposedStore for input operations
func InputStoreComposed(ctx context.Context, walletType string) (*ComposedStore, error) {
	return newComposedStore(ctx, "input", walletType)
}

// OutputStoreComposed creates a new ComposedStore for output operations
func OutputStoreComposed(ctx context.Context, walletType string) (*ComposedStore, error) {
	return newComposedStore(ctx, "output", walletType)
}

func newComposedStore(ctx context.Context, side string, walletType string) (*ComposedStore, error) {
	switch walletType {
	case "hierarchical deterministic":
		s, err := newHDStore(ctx, side)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create HD store")
		}
		return NewComposedStore([]AtomicStore{s}, "HD"), nil

	case "non-deterministic":
		s, err := newSimpleStore(ctx, side, "ND Store")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ND store")
		}
		return NewComposedStore([]AtomicStore{s}, "ND"), nil

	case "keystore":
		s, err := newSimpleStore(ctx, side, "Keystore Store")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Keystore store")
		}
		return NewComposedStore([]AtomicStore{s}, "Keystore"), nil

	case "distributed":
		atomics, err := newDistributedAtomicStores(ctx, side)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Distributed stores")
		}
		return NewComposedStore(atomics, "Distributed"), nil

	default:
		return nil, fmt.Errorf("incorrect %s wallet type: %q", side, walletType)
	}
}

// newDistributedAtomicStores creates one atomic store per configured peer
func newDistributedAtomicStores(ctx context.Context, side string) ([]AtomicStore, error) {
	walletType := viper.GetString(side + ".wallet.type")
	storePath := viper.GetString(side + ".store.path")
	if storePath == "" {
		return nil, errors.New("distributed store path is empty")
	}

	var peers map[uint64]Peer
	if err := viper.UnmarshalKey(side+".wallet.peers", &peers); err != nil {
		return nil, err
	}
	if len(peers) < 2 {
		return nil, errors.New("number of peers for distributed store is less than 2")
	}

	threshold := viper.GetUint32(side + ".wallet.threshold")
	if threshold <= uint32(len(peers)/2) || threshold > uint32(len(peers)) {
		return nil, errors.New("invalid threshold value for distributed store")
	}

	// Build the full peer map first: every atomic store gets the complete map
	peersMap := make(map[uint64]string, len(peers))
	for id, peer := range peers {
		peersMap[id] = peer.Name
	}

	atomicStores := make([]AtomicStore, 0, len(peers))
	for id, peer := range peers {
		passphrases, err := getAccountsPasswords(peer.Passphrases.Path)
		if err != nil {
			return nil, err
		}
		passphrases, err = selectPassphrase(passphrases, fmt.Sprintf("%s.wallet.peers.%d.passphrases.index", side, id))
		if err != nil {
			return nil, err
		}

		// Peer name is "host:port"; the peer directory is named after the host
		peerDir, _, _ := strings.Cut(peer.Name, ":")
		peerPath := storePath + "/" + peerDir

		atomicStore, err := newDistributedAtomicStore(ctx, side, id, peer.Name, peerPath, walletType, passphrases, threshold, peersMap)
		if err != nil {
			return nil, err
		}

		atomicStores = append(atomicStores, atomicStore)
	}

	return atomicStores, nil
}
