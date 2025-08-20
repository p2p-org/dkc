package store

import (
	"bytes"
	"context"
	"os"

	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type AccountsData struct {
	Name  string
	WName string
}

func lockAccount(ctx context.Context, acc types.Account) error {
	if locker, isLocker := acc.(types.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return err
		}
		if unlocked {
			err := locker.Lock(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func unlockAccount(ctx context.Context, acc types.Account, passphrases [][]byte) (types.Account, error) {
	account := acc
	if locker, isLocker := account.(types.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return nil, err
		}
		if !unlocked {
			for _, passphrase := range passphrases {
				err = locker.Unlock(ctx, passphrase)
				if err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				err = errors.Wrap(nil, "failed to ulock account")
				return nil, err
			}
		}
	}

	return account, nil
}

func getAccountsPasswords(path string) ([][]byte, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	accountsPasswords := bytes.Split(content, []byte{'\n'})
	if len(accountsPasswords) == 0 {
		return nil, errors.New("accounts passwords is empty")
	}
	return accountsPasswords, nil
}

func createStore(path string) (types.Store, error) {
	store := filesystem.New(filesystem.WithLocation(path))
	return store, nil
}

func getStore(path string) (types.Store, error) {
	store := filesystem.New(filesystem.WithLocation(path))
	return store, nil
}

func getWalletsAccountsMap(ctx context.Context, location string) ([]AccountsData, []string, error) {
	accs := []AccountsData{}
	wallets := []string{}
	store := filesystem.New(filesystem.WithLocation(location))
	if err := e2wallet.UseStore(store); err != nil {
		return accs, wallets, err
	}
	for w := range e2wallet.Wallets() {
		wallets = append(wallets, w.Name())
		for a := range w.Accounts(ctx) {
			accs = append(accs, AccountsData{Name: a.Name(), WName: w.Name()})
		}
	}

	return accs, wallets, nil
}

func getAccountPrivateKey(ctx context.Context, account types.Account, passphrases [][]byte) ([]byte, error) {
	privateKeyProvider, isPrivateKeyProvider := account.(types.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		err := errors.New("failed to get account method")
		return nil, err
	}

	_, err := unlockAccount(ctx, account, passphrases)
	if err != nil {
		return nil, err
	}

	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		return nil, err
	}

	// Lock Account
	defer func() {
		err = lockAccount(ctx, account)
	}()

	return key.Marshal(), nil
}

func getWallet(location string, n string) (types.Wallet, error) {
	store := filesystem.New(filesystem.WithLocation(location))
	if err := e2wallet.UseStore(store); err != nil {
		return nil, err
	}
	w, err := e2wallet.OpenWallet(n)
	if err != nil {
		return nil, err
	}

	return w, nil
}
