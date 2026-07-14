package store

import (
	"context"
	"runtime"
	"sync"

	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// WalletCache provides fast in-memory access to wallets and accounts
type WalletCache struct {
	mu             sync.RWMutex
	wallets        map[string]types.Wallet
	walletAccounts map[string]map[string]types.Account
	initialized    bool
}

// accountResult holds the result of processing an account
type accountResult struct {
	account     types.Account
	accountName string
	unlocked    bool
	err         error
}

// NewWalletCache creates a new wallet cache
func NewWalletCache() *WalletCache {
	return &WalletCache{
		wallets:        make(map[string]types.Wallet),
		walletAccounts: make(map[string]map[string]types.Account),
	}
}

// PopulateFromLocation loads all wallets and accounts from a location into
// cache. logPrefix (may be empty) tags every log line, e.g. "Peer 10"
func (wc *WalletCache) PopulateFromLocation(ctx context.Context, location string, passphrases [][]byte, logPrefix string) error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	log := utils.Log
	if logPrefix != "" {
		log = log.With().Str("peer", logPrefix).Logger()
	}

	if wc.initialized {
		log.Debug().Msgf("cache already initialized for location: %s", location)
		return nil
	}

	log.Info().Msgf("🔄 starting wallet cache population from location: %s", location)

	store := newFSStore(location)

	walletCount := 0
	accountCount := 0
	unlockedCount := 0

	for wallet := range e2wallet.Wallets(e2wallet.WithStore(store)) {
		walletName := wallet.Name()
		log.Info().Msgf("📁 processing wallet: %s", walletName)

		wc.wallets[walletName] = wallet
		wc.walletAccounts[walletName] = make(map[string]types.Account)
		walletCount++

		// Collect all accounts from this wallet first
		var accounts []types.Account
		for account := range wallet.Accounts(ctx) {
			accounts = append(accounts, account)
		}

		if len(accounts) == 0 {
			log.Info().Msgf("📭 wallet %s has no accounts", walletName)
			continue
		}

		// Process accounts using worker pool for parallel unlock
		numWorkers := min(runtime.NumCPU()*3, len(accounts))
		log.Info().Msgf("🔧 processing %d accounts with %d workers for wallet: %s", len(accounts), numWorkers, walletName)

		accountChan := make(chan types.Account, len(accounts))
		resultChan := make(chan accountResult, len(accounts))

		var workersWg sync.WaitGroup
		for i := range numWorkers {
			workersWg.Add(1)
			go func(workerID int) {
				defer workersWg.Done()
				for account := range accountChan {
					resultChan <- processAccount(ctx, account, walletName, passphrases, log, workerID)
				}
			}(i)
		}

		for _, account := range accounts {
			accountChan <- account
		}
		close(accountChan)

		workersWg.Wait()
		close(resultChan)

		walletAccountCount := 0
		for result := range resultChan {
			if result.err != nil {
				log.Warn().Err(result.err).Msgf("⚠️ warning processing account: %s/%s", walletName, result.accountName)
			}

			wc.walletAccounts[walletName][result.accountName] = result.account
			accountCount++
			walletAccountCount++

			if result.unlocked {
				unlockedCount++
			}
		}
		log.Info().Msgf("✅ wallet %s: %d accounts cached", walletName, walletAccountCount)
	}

	wc.initialized = true

	log.Info().Msgf("🎉 cache population completed: %d wallets, %d accounts, %d unlocked (location: %s)",
		walletCount, accountCount, unlockedCount, location)

	return nil
}

// GetOrOpenWallet returns a cached wallet instance, opening and caching it on
// first use. All writers share one instance per wallet, so concurrent imports
// serialize on the wallet's own mutex and cannot lose index updates.
func (wc *WalletCache) GetOrOpenWallet(location string, walletName string) (types.Wallet, error) {
	wc.mu.RLock()
	wallet, exists := wc.wallets[walletName]
	wc.mu.RUnlock()
	if exists {
		return wallet, nil
	}

	wc.mu.Lock()
	defer wc.mu.Unlock()
	if wallet, exists := wc.wallets[walletName]; exists {
		return wallet, nil
	}
	wallet, err := getWallet(location, walletName)
	if err != nil {
		return nil, err
	}
	wc.wallets[walletName] = wallet
	return wallet, nil
}

// PutWallet caches an already-open wallet instance (e.g. right after creation)
func (wc *WalletCache) PutWallet(wallet types.Wallet) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.wallets[wallet.Name()] = wallet
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

// processAccount handles unlocking a single account in a worker
func processAccount(ctx context.Context, account types.Account, walletName string, passphrases [][]byte, log zerolog.Logger, workerID int) accountResult {
	accountName := account.Name()
	result := accountResult{
		account:     account,
		accountName: accountName,
	}

	log.Debug().Msgf("👤 worker %d processing account: %s/%s", workerID, walletName, accountName)

	locker, isLocker := account.(types.AccountLocker)
	if !isLocker {
		return result
	}

	unlocked, err := locker.IsUnlocked(ctx)
	if err != nil {
		result.err = err
		return result
	}
	if unlocked {
		log.Debug().Msgf("✅ worker %d account %s/%s already unlocked", workerID, walletName, accountName)
		result.unlocked = true
		return result
	}

	for i, passphrase := range passphrases {
		if err := locker.Unlock(ctx, passphrase); err == nil {
			log.Debug().Msgf("🔓 worker %d unlocked account %s/%s with passphrase #%d", workerID, walletName, accountName, i)
			result.unlocked = true
			break
		}
	}

	return result
}
