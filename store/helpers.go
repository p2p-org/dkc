package store

import (
	"bytes"
	"context"
	"os"
	"sync"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type AccountsData struct {
	Name  string
	WName string
}

// Global file mutex map to prevent concurrent wallet file corruption
var (
	fileMutexes   = make(map[string]*sync.Mutex)
	fileMutexesMu sync.RWMutex
)

// getFileMutex returns a mutex for a specific wallet path to prevent file corruption
func getFileMutex(walletPath string) *sync.Mutex {
	fileMutexesMu.RLock()
	mu, exists := fileMutexes[walletPath]
	fileMutexesMu.RUnlock()

	if exists {
		return mu
	}

	fileMutexesMu.Lock()
	defer fileMutexesMu.Unlock()

	// Double-check in case another goroutine created it
	if mu, exists := fileMutexes[walletPath]; exists {
		return mu
	}

	mu = &sync.Mutex{}
	fileMutexes[walletPath] = mu
	return mu
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

	// REMOVED: defer lockAccount - keep accounts unlocked for wallet operations
	// This was causing "wallet must be unlocked to import accounts" errors

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

	// Always unlock wallet after opening (follow "unlock once, never lock" strategy)
	if locker, isLocker := w.(types.WalletLocker); isLocker {
		// Check if already unlocked first
		unlocked, err := locker.IsUnlocked(context.Background())
		if err == nil && !unlocked {
			// Unlock with nil passphrase (works for ND wallets)
			err = locker.Unlock(context.Background(), nil)
			if err != nil {
				utils.Log.Warn().Err(err).Msgf("⚠️ Failed to unlock wallet after opening: %s", n)
				// Don't return error - some wallets might not need unlocking
			} else {
				utils.Log.Debug().Msgf("🔓 Wallet %s unlocked after opening from disk", n)
			}
		}
	}

	return w, nil
}
