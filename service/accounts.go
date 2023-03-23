package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func CreateNDAccount(
	key []byte,
	name string,
	passphrase []byte,
	wallet types.Wallet,
) {
	err := wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer wallet.(types.WalletLocker).Lock(context.Background())

	account, err := wallet.(types.WalletAccountImporter).ImportAccount(context.Background(),
		name,
		key,
		passphrase,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(account)
}

func CreateAccounts(
	wallet types.Wallet,
	accountsPassword []byte,
	name string,
	masterPKs [][]byte,
	masterSKs [][]byte,
	threshold uint32,
	peers map[uint64]string,
) {
	err := wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	// Always immediately defer locking the wallet to ensure it does not remain unlocked outside of the function.
	defer wallet.(types.WalletLocker).Lock(context.Background())

	privateKey := masterSKs[0]
	signingThreshold := threshold
	verificationVector := masterPKs
	participants := peers

	account, err := wallet.(types.WalletDistributedAccountImporter).ImportDistributedAccount(context.Background(),
		name,
		privateKey,
		signingThreshold,
		verificationVector,
		participants,
		accountsPassword)
	if err != nil {
		panic(err)
	}

	fmt.Println(account)

	return

}

func GetAccountKey(ctx context.Context, account types.Account, passphrases [][]byte) ([]byte, error) {
	privateKeyProvider, isPrivateKeyProvider := account.(types.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		fmt.Println("account does not provide its private key")
	}

	if locker, isLocker := account.(types.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find out if account is locked")
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
