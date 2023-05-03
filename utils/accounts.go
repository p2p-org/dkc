package utils

import (
	"context"
	"fmt"

	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type ndWallet interface {
	types.WalletAccountImporter
}

type dWallet interface {
	types.WalletDistributedAccountImporter
}

const signingString = "bkeCE2vRuTxxc5RpzrvLzoU5EgulV7uk3zMnt5MP9MgsXBaif9mUQcf7rZGC5mNj9lBqQ2s"

func CreateAccount(
	wallet types.Wallet,
	name string,
	masterPKs [][]byte,
	masterSK []byte,
	threshold uint32,
	peers map[uint64]string,
	passphrase []byte,
) (account types.Account) {

	err := wallet.(types.WalletLocker).Unlock(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	defer wallet.(types.WalletLocker).Lock(context.Background())

	switch wallet := wallet.(type) {
	case ndWallet:
		account, err = wallet.ImportAccount(context.Background(),
			name,
			masterSK,
			passphrase,
		)
		if err != nil {
			panic(err)
		}
	case dWallet:
		signingThreshold := threshold
		verificationVector := masterPKs
		participants := peers

		account, err = wallet.ImportDistributedAccount(context.Background(),
			name,
			masterSK,
			signingThreshold,
			verificationVector,
			participants,
			passphrase)
		if err != nil {
			panic(err)
		}
	default:
		panic(err)
	}

	return
}

func AccountSign(ctx context.Context, acc types.Account, passphrases [][]byte) []byte {
	account := unlockAccount(ctx, acc, passphrases)
	accountSigner := account.(types.AccountSigner)

	signedData, err := accountSigner.Sign(ctx, []byte(signingString))
	if err != nil {
		panic(err)
	}

	if !signedData.Verify([]byte(signingString), acc.PublicKey()) {
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

func GetAccountCompositePubkey(account types.Account) ([]byte, error) {
	compositePublicKeyProvider, isCompositePublicKeyProvider := account.(types.AccountCompositePublicKeyProvider)
	if !isCompositePublicKeyProvider {
		fmt.Println("account does not provide its private key")
	}

	pubKey := compositePublicKeyProvider.CompositePublicKey()

	return pubKey.Marshal(), nil
}

func GetAccountPubkey(account types.Account) ([]byte, error) {
	return account.PublicKey().Marshal(), nil
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
