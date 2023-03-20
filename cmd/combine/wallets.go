package combine

import (
	"context"
	"fmt"
	"io/ioutil"

	// "log"
	"strings"
	// "encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	// e2wallet "github.com/wealdtech/go-eth2-wallet"
	// filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type WalletName = string
type AccountName = string
type Key = []byte
type KeyMapping = map[WalletName]map[AccountName][]Key

func getAccountKey(ctx context.Context, account e2wtypes.Account) (Key, error) {
	passphrases_path := viper.GetString("passphrases")

	if passphrases_path == "" {
		return nil, errors.New("No passphrase file")
	}
	content, err := ioutil.ReadFile(passphrases_path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read passphrase file")
	}

	passphrases := strings.Split(string(content), "\n")
	privateKeyProvider, isPrivateKeyProvider := account.(e2wtypes.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		fmt.Println("account does not provide its private key")
	}

	if locker, isLocker := account.(e2wtypes.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find out if account is locked")
		}
		if !unlocked {
			for _, passphrase := range passphrases {
				err = locker.Unlock(ctx, []byte(passphrase))
				if err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				return nil, errors.New("failed to unlock account")
			}
		}
	}
	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		fmt.Println(err, "failed to obtain private key")
	}
	return key.Marshal(), nil
}

// func getWalletKeys(ctx context.Context, location string) {
//   keys := make(map[WalletName]map[AccountName][]Key)
//
// 	store := filesystem.New(filesystem.WithLocation(location))
// 	if err := e2wallet.UseStore(store); err != nil {
// 		fmt.Println("failed to UseStore: %w", err)
// 	}
// 	for wallet := range e2wallet.Wallets() {
//     keys[wallet.Name()] = make(map[string][]Key)
//
// 		for account := range wallet.Accounts(ctx) {
//       key, err := getAccountKey(ctx, account)
//       if err != nil {
//         fmt.Println("Error")
//       }
//       keys[wallet.Name()][account.Name()] = append(keys[wallet.Name()][account.Name()], key)
// 		}
// 	}
//   bs, _ := json.Marshal(keys)
//   println("Keys:", string(bs))
// }

// func loadWallet(ctx context.Context, location string) {
// 	store := filesystem.New(filesystem.WithLocation(location))
// 	if err := e2wallet.UseStore(store); err != nil {
// 		fmt.Println("failed to UseStore: %w", err)
// 	}
// 	for wallet := range e2wallet.Wallets() {
//     keys[wallet.Name()] = make(map[string][]Key)
//
// 		for account := range wallet.Accounts(ctx) {
//       key, err := getAccountKey(ctx, account)
//       if err != nil {
//         fmt.Println("Error")
//       }
//       keys[wallet.Name()][account.Name()] = append(keys[wallet.Name()][account.Name()], key)
// 		}
// 	}
//   bs, _ := json.Marshal(keys)
//   println("Keys:", string(bs))
// }
