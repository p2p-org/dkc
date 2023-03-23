package split

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type Peers map[uint64]string

func getAccountsPassword() []byte {
	accountsPasswordPath := viper.GetString("passphrases")

	accountsPassword, err := os.ReadFile(accountsPasswordPath)
	if err != nil {
		panic(err)
	}

	return accountsPassword
}

func getMasterKey() []byte {
	masterKeyPath := viper.GetString("master-key")

	masterKey, err := os.ReadFile(masterKeyPath)
	if err != nil {
		panic(err)
	}

	return masterKey
}

func createWallet(store types.Store) (wallet types.Wallet) {
	e2wallet.UseStore(store)
	wallet, err := e2wallet.CreateWallet(uuid.New().String(), e2wallet.WithType("distributed"))
	if err != nil {
		panic(err)
	}
	return
}

func createAccounts(
	wallet types.Wallet,
	accountsPassword []byte,
	masterPKs [][]byte,
	masterSKs [][]byte,
	threshold uint32,
	peers Peers,
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

func CreateWallets(ctx context.Context) error {
	threshold := viper.GetUint32("signing-threshold")
	masterKey := getMasterKey()
	accountsPassword := getAccountsPassword()
	var peers Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	masterSKs, masterPKs := bls.Split(ctx, masterKey, threshold)

	store := createStore("./newwallets")
	wallet := createWallet(store)

	createAccounts(
		wallet,
		accountsPassword,
		masterPKs,
		masterSKs,
		threshold,
		peers,
	)

	fmt.Println(wallet)
	return nil
}
