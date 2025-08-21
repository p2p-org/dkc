package store

import (
	"context"
	"sync"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// WalletCache provides fast in-memory access to wallets and accounts
type WalletCache struct {
	mu             sync.RWMutex
	wallets        map[string]types.Wallet
	walletAccounts map[string]map[string]types.Account
	pubKeyPaths    map[[48]byte]string
	initialized    bool
}

// NewWalletCache creates a new wallet cache
func NewWalletCache() *WalletCache {
	return &WalletCache{
		wallets:        make(map[string]types.Wallet),
		walletAccounts: make(map[string]map[string]types.Account),
		pubKeyPaths:    make(map[[48]byte]string),
		initialized:    false,
	}
}

// PopulateFromLocation loads all wallets and accounts from a location into cache
func (wc *WalletCache) PopulateFromLocation(ctx context.Context, location string, passphrases [][]byte) error {
	return wc.PopulateFromLocationWithPrefix(ctx, location, passphrases, "")
}

// PopulateFromLocationWithPrefix loads all wallets and accounts from a location into cache with custom log prefix
func (wc *WalletCache) PopulateFromLocationWithPrefix(ctx context.Context, location string, passphrases [][]byte, logPrefix string) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if wc.initialized {
		if logPrefix != "" {
			utils.Log.Debug().Msgf("cache already initialized for location: %s [%s]", location, logPrefix)
		} else {
			utils.Log.Debug().Msgf("cache already initialized for location: %s", location)
		}
		return nil // Already initialized
	}

	if logPrefix != "" {
		utils.Log.Info().Msgf("🔄 starting wallet cache population from location: %s [%s]", location, logPrefix)
		utils.Log.Info().Msgf("💾 initializing filesystem store for location: %s [%s]", location, logPrefix)
	} else {
		utils.Log.Info().Msgf("🔄 starting wallet cache population from location: %s", location)
		utils.Log.Info().Msgf("💾 initializing filesystem store for location: %s", location)
	}

	store := filesystem.New(filesystem.WithLocation(location))
	if err := e2wallet.UseStore(store); err != nil {
		return err
	}

	walletCount := 0
	accountCount := 0
	unlockedCount := 0
	pubKeyCount := 0

	// Process each wallet
	if logPrefix != "" {
		utils.Log.Info().Msgf("🔍 discovering wallets in location: %s [%s]", location, logPrefix)
	} else {
		utils.Log.Info().Msgf("🔍 discovering wallets in location: %s", location)
	}

	for wallet := range e2wallet.Wallets() {
		walletName := wallet.Name()
		if logPrefix != "" {
			utils.Log.Info().Msgf("📁 processing wallet: %s [%s]", walletName, logPrefix)
		} else {
			utils.Log.Info().Msgf("📁 processing wallet: %s", walletName)
		}

		wc.wallets[walletName] = wallet
		wc.walletAccounts[walletName] = make(map[string]types.Account)
		walletCount++

		walletAccountCount := 0
		// Process accounts in this wallet
		for account := range wallet.Accounts(ctx) {
			accountName := account.Name()
			if logPrefix != "" {
				utils.Log.Debug().Msgf("👤 processing account: %s/%s [%s]", walletName, accountName, logPrefix)
			} else {
				utils.Log.Debug().Msgf("👤 processing account: %s/%s", walletName, accountName)
			}

			// Unlock account with provided passphrases
			if locker, isLocker := account.(types.AccountLocker); isLocker {
				unlocked, err := locker.IsUnlocked(ctx)
				if err == nil && !unlocked {
					for i, passphrase := range passphrases {
						if err := locker.Unlock(ctx, passphrase); err == nil {
							if logPrefix != "" {
								utils.Log.Debug().Msgf("🔓 unlocked account %s/%s with passphrase #%d [%s]", walletName, accountName, i, logPrefix)
							} else {
								utils.Log.Debug().Msgf("🔓 unlocked account %s/%s with passphrase #%d", walletName, accountName, i)
							}
							unlockedCount++
							break
						}
					}
				} else if err == nil && unlocked {
					if logPrefix != "" {
						utils.Log.Debug().Msgf("✅ account %s/%s already unlocked [%s]", walletName, accountName, logPrefix)
					} else {
						utils.Log.Debug().Msgf("✅ account %s/%s already unlocked", walletName, accountName)
					}
					unlockedCount++
				}
			}

			wc.walletAccounts[walletName][accountName] = account
			accountCount++
			walletAccountCount++

			// Store public key mapping if available
			if pubKeyProvider, ok := account.(types.AccountPublicKeyProvider); ok {
				if pubKey := pubKeyProvider.PublicKey(); pubKey != nil {
					var pubKeyBytes [48]byte
					copy(pubKeyBytes[:], pubKey.Marshal())
					wc.pubKeyPaths[pubKeyBytes] = walletName + "/" + accountName
					pubKeyCount++
					if logPrefix != "" {
						utils.Log.Debug().Msgf("🔑 indexed public key for account: %s/%s [%s]", walletName, accountName, logPrefix)
					} else {
						utils.Log.Debug().Msgf("🔑 indexed public key for account: %s/%s", walletName, accountName)
					}
				}
			}
		}
		if logPrefix != "" {
			utils.Log.Info().Msgf("✅ wallet %s: %d accounts cached [%s]", walletName, walletAccountCount, logPrefix)
		} else {
			utils.Log.Info().Msgf("✅ wallet %s: %d accounts cached", walletName, walletAccountCount)
		}
	}

	wc.initialized = true

	if logPrefix != "" {
		utils.Log.Info().Msgf("🎉 cache population completed! [%s]", logPrefix)
		utils.Log.Info().Msgf("📊 cache stats: %d wallets, %d accounts, %d unlocked, %d public keys indexed [%s]",
			walletCount, accountCount, unlockedCount, pubKeyCount, logPrefix)
		utils.Log.Info().Msgf("📍 location: %s [%s]", location, logPrefix)
	} else {
		utils.Log.Info().Msgf("🎉 cache population completed!")
		utils.Log.Info().Msgf("📊 cache stats: %d wallets, %d accounts, %d unlocked, %d public keys indexed",
			walletCount, accountCount, unlockedCount, pubKeyCount)
		utils.Log.Info().Msgf("📍 location: %s", location)
	}

	return nil
}

// FetchWallet retrieves a wallet from cache
func (wc *WalletCache) FetchWallet(walletName string) (types.Wallet, error) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	utils.Log.Debug().Msgf("🔍 fetching wallet from cache: %s", walletName)

	wallet, exists := wc.wallets[walletName]
	if !exists {
		utils.Log.Warn().Msgf("❌ wallet not found in cache: %s", walletName)
		return nil, errors.New("wallet not found in cache")
	}

	utils.Log.Debug().Msgf("✅ wallet found in cache: %s", walletName)
	return wallet, nil
}

// FetchAccount retrieves an account from cache
func (wc *WalletCache) FetchAccount(walletName, accountName string) (types.Account, error) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	utils.Log.Debug().Msgf("🔍 fetching account from cache: %s/%s", walletName, accountName)

	accounts, exists := wc.walletAccounts[walletName]
	if !exists {
		utils.Log.Warn().Msgf("❌ wallet not found in cache: %s", walletName)
		return nil, errors.New("wallet not found in cache")
	}

	account, exists := accounts[accountName]
	if !exists {
		utils.Log.Warn().Msgf("❌ account not found in cache: %s/%s", walletName, accountName)
		return nil, errors.New("account not found in cache")
	}

	utils.Log.Debug().Msgf("✅ account found in cache: %s/%s", walletName, accountName)
	return account, nil
}

// FetchAccountByPublicKey retrieves an account by its public key
func (wc *WalletCache) FetchAccountByPublicKey(pubKey []byte) (types.Account, error) {
	if len(pubKey) != 48 {
		utils.Log.Warn().Msgf("❌ invalid public key length: %d, expected 48", len(pubKey))
		return nil, errors.New("invalid public key length")
	}

	var pubKeyBytes [48]byte
	copy(pubKeyBytes[:], pubKey)

	utils.Log.Debug().Msgf("🔍 fetching account by public key from cache")

	wc.mu.RLock()
	defer wc.mu.RUnlock()

	path, exists := wc.pubKeyPaths[pubKeyBytes]
	if !exists {
		utils.Log.Warn().Msgf("❌ account not found by public key in cache")
		return nil, errors.New("account not found by public key")
	}

	utils.Log.Debug().Msgf("✅ found account path by public key: %s", path)

	// Parse wallet and account names from path
	walletName, accountName, err := e2wallet.WalletAndAccountNames(path)
	if err != nil {
		utils.Log.Error().Err(err).Msgf("❌ invalid cached path: %s", path)
		return nil, errors.Wrap(err, "invalid cached path")
	}

	return wc.FetchAccount(walletName, accountName)
}

// GetWalletsAccountsMap returns all accounts and wallet names for compatibility
func (wc *WalletCache) GetWalletsAccountsMap() ([]AccountsData, []string) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	var accounts []AccountsData
	var wallets []string

	for walletName := range wc.wallets {
		wallets = append(wallets, walletName)
		if accountsMap, exists := wc.walletAccounts[walletName]; exists {
			for accountName := range accountsMap {
				accounts = append(accounts, AccountsData{
					Name:  accountName,
					WName: walletName,
				})
			}
		}
	}

	return accounts, wallets
}

// IsInitialized returns whether the cache has been populated
func (wc *WalletCache) IsInitialized() bool {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.initialized
}
