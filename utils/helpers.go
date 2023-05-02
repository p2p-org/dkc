package utils

import (
	"bytes"
	"os"

	"github.com/spf13/viper"
)

func GetAccountsPasswords() [][]byte {
	accountsPasswordPath := viper.GetString("passphrases")

	content, err := os.ReadFile(accountsPasswordPath)
	if err != nil {
		panic(err)
	}

	accountsPasswords := bytes.Split(content, []byte{'\n'})
	return accountsPasswords
}
