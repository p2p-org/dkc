package service

import (
	"context"
	"fmt"

	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func CreateNDAccount(
	key []byte,
	name string,
	passphrase []byte,
	wallet types.Wallet,
) (account types.Account) {
	err := wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	defer wallet.(types.WalletLocker).Lock(context.Background())

	account, err = wallet.(types.WalletAccountImporter).ImportAccount(context.Background(),
		name,
		key,
		passphrase,
	)
	if err != nil {
		panic(err)
	}

	return account
}

func CreateAccount(
	wallet types.Wallet,
	accountsPassword []byte,
	name string,
	masterPKs [][]byte,
	masterSK []byte,
	threshold uint32,
	peers map[uint64]string,
) (account types.Account) {
	err := wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	// Always immediately defer locking the wallet to ensure it does not remain unlocked outside of the function.
	defer wallet.(types.WalletLocker).Lock(context.Background())

	signingThreshold := threshold
	verificationVector := masterPKs
	participants := peers

	account, err = wallet.(types.WalletDistributedAccountImporter).ImportDistributedAccount(context.Background(),
		name,
		masterSK,
		signingThreshold,
		verificationVector,
		participants,
		accountsPassword)
	if err != nil {
		panic(err)
	}

	return

}

func AccountSign(ctx context.Context, acc types.Account, signerData []byte, passphrases [][]byte) []byte {
	account := unlockAccount(ctx, acc, passphrases)
	accountSigner := account.(types.AccountSigner)

	signedData, err := accountSigner.Sign(ctx, signerData)
	if err != nil {
		panic(err)
	}

	if !signedData.Verify(signerData, acc.PublicKey()) {
		panic("rap")
	}

	return signedData.Marshal()
}

func GetAccountKey(ctx context.Context, account types.Account, passphrases [][]byte) ([]byte, error) {
	privateKeyProvider, isPrivateKeyProvider := account.(types.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		fmt.Println("account does not provide its private key")
	}

	unlockAccount(ctx, account, passphrases)

	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		fmt.Println(err, "failed to obtain private key")
	}

	return key.Marshal(), nil
}

func unlockAccount(ctx context.Context, acc types.Account, passphrases [][]byte) (account types.Account) {
	account = acc
	if locker, isLocker := account.(types.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			panic(err)
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
				panic(err)
			}
		}
	}

	return
}
