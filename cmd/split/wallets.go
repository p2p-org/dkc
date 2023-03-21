package split

import (
	"context"
	"fmt"
	"os"

	"github.com/p2p-org/dkc/service/crypto/bls"
	"github.com/spf13/viper"
)

type Peers map[uint64]string

func getMasterKey() []byte {
	masterKeyPath := viper.GetString("master-key")

	masterKey, err := os.ReadFile(masterKeyPath)
	if err != nil {
		panic(err)
	}

	return masterKey
}

func CreateWallets(ctx context.Context) error {
	//passphrasesPath := viper.GetString("passphrases")
	masterKey := getMasterKey()
	threshold := viper.GetInt("signing-threshold")
	peersIDs := []uint64{}
	var peers Peers
	err := viper.UnmarshalKey("peers", &peers)
	if err != nil {
		fmt.Println(err)
	}

	for key := range peers {
		peersIDs = append(peersIDs, key)
	}

	bls.Split(ctx, masterKey, peersIDs, threshold)

	return nil
}
