package store

import (
	"context"

	"github.com/pkg/errors"
)

type IStore interface {
	// Get Wallets Names And Account Names
	GetWalletsAccountsMap() ([]AccountsData, []string, error)
	// Get Private Key From Wallet Using Account Name
	GetPK(a string, w string) ([]byte, error)
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
}

type OStore interface {
	// Create Store
	Create() error
	// Create New Wallet
	CreateWallet(name string) error
	// Save Private Key To Wallet
	SavePKToWallet(w string, a []byte, n string) error
	// Get Store Type
	GetType() string
	// Get Store Path
	GetPath() string
}

func InputStoreInit(ctx context.Context, t string) (IStore, error) {
	switch t {
	case "distributed":
		s, err := newDistributedStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as input store")
		}
		s.Ctx = ctx
		return &s, nil
	case "hierarchical deterministic":
		s, err := newHDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init hierarchial deterministic store as input store")
		}
		s.Ctx = ctx
		return &s, nil
	case "non-deterministic":
		s, err := newNDStore("input")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as input store")
		}
		s.Ctx = ctx
		return &s, nil
	default:
		return nil, errors.New("incorrect input wallet type")
	}

}

func OutputStoreInit(ctx context.Context, t string) (OStore, error) {
	switch t {
	case "distributed":
		s, err := newDistributedStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init distributed store as output store")
		}
		s.Ctx = ctx
		return &s, nil
	case "non-deterministic":
		s, err := newNDStore("output")
		if err != nil {
			return nil, errors.Wrap(err, "failed to init non-deterministic store as output store")
		}
		s.Ctx = ctx
		return &s, nil
	default:
		return nil, errors.New("incorrect output wallet type")
	}

}
