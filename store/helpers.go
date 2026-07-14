package store

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type AccountsData struct {
	Name  string
	WName string
}

// storeConfig is the common per-side ("input"/"output") store configuration
type storeConfig struct {
	WalletType  string
	Path        string
	Passphrases [][]byte
}

func parseStoreConfig(side string) (storeConfig, error) {
	cfg := storeConfig{}

	cfg.WalletType = viper.GetString(side + ".wallet.type")
	utils.Log.Debug().Msgf("setting store type to %s", cfg.WalletType)

	cfg.Path = viper.GetString(side + ".store.path")
	utils.Log.Debug().Msgf("setting store path to %s", cfg.Path)
	if cfg.Path == "" {
		return cfg, fmt.Errorf("%s store path is empty", side)
	}

	passphrases, err := getAccountsPasswords(viper.GetString(side + ".wallet.passphrases.path"))
	if err != nil {
		return cfg, err
	}
	cfg.Passphrases, err = selectPassphrase(passphrases, side+".wallet.passphrases.index")
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getAccountsPasswords(path string) ([][]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	accountsPasswords := [][]byte{}
	for passphrase := range bytes.SplitSeq(content, []byte{'\n'}) {
		if len(passphrase) > 0 {
			accountsPasswords = append(accountsPasswords, passphrase)
		}
	}
	if len(accountsPasswords) == 0 {
		return nil, errors.New("accounts passwords is empty")
	}
	return accountsPasswords, nil
}

// selectPassphrase narrows passphrases down to a single entry when an index
// is configured at the given viper key
func selectPassphrase(passphrases [][]byte, indexKey string) ([][]byte, error) {
	if !viper.IsSet(indexKey) {
		return passphrases, nil
	}
	index := viper.GetInt(indexKey)
	if index < 0 || index >= len(passphrases) {
		return nil, fmt.Errorf("passphrases index %d is out of range: %d passphrases available", index, len(passphrases))
	}
	return [][]byte{passphrases[index]}, nil
}

func newFSStore(path string) types.Store {
	return filesystem.New(filesystem.WithLocation(path))
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
				if err := locker.Unlock(ctx, passphrase); err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				return nil, errors.New("failed to unlock account: no passphrase matched")
			}
		}
	}

	return account, nil
}

func getWalletsAccountsMap(ctx context.Context, location string) ([]AccountsData, []string, error) {
	accs := []AccountsData{}
	wallets := []string{}
	for w := range e2wallet.Wallets(e2wallet.WithStore(newFSStore(location))) {
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
		return nil, errors.New("account does not provide its private key")
	}

	if _, err := unlockAccount(ctx, account, passphrases); err != nil {
		return nil, err
	}

	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		return nil, err
	}

	// Keep the account unlocked: locking it back would break subsequent
	// wallet operations that expect unlocked accounts

	return key.Marshal(), nil
}

func getWallet(location string, n string) (types.Wallet, error) {
	w, err := e2wallet.OpenWallet(n, e2wallet.WithStore(newFSStore(location)))
	if err != nil {
		return nil, err
	}

	// Always unlock wallet after opening (follow "unlock once, never lock" strategy)
	if locker, isLocker := w.(types.WalletLocker); isLocker {
		unlocked, err := locker.IsUnlocked(context.Background())
		if err == nil && !unlocked {
			// Unlock with nil passphrase (works for ND wallets)
			if err := locker.Unlock(context.Background(), nil); err != nil {
				// Don't return error - some wallets might not need unlocking
				utils.Log.Warn().Err(err).Msgf("⚠️ failed to unlock wallet after opening: %s", n)
			} else {
				utils.Log.Debug().Msgf("🔓 wallet %s unlocked after opening from disk", n)
			}
		}
	}

	return w, nil
}
