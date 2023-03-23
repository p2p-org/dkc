package split

import (
	"os"

	"github.com/spf13/viper"
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
