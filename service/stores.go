package service

import (
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func createStore(path string) (store types.Store) {
	store = filesystem.New(filesystem.WithLocation(path))
	return
}
