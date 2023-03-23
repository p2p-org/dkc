package service

import (
	"context"

	"github.com/google/uuid"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func createAccounts(
	wallet types.Wallet,
	accountsPassword []byte,
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
		uuid.New().String(),
		privateKey,
		signingThreshold,
		verificationVector,
		participants,
		accountsPassword)
	if err != nil {
		panic(err)
	}

	print(account)
	return

}
