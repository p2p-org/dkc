package utils

import (
	"bytes"
	"os"
)

type NDWalletConfig struct {
	Path        string
	Passphrases string
}

type DWalletConfig struct {
	Path        string
	Passphrases string
	Peers       Peers
	Threshold   uint32
}

func GetAccountsPasswords(path string) ([][]byte, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	accountsPasswords := bytes.Split(content, []byte{'\n'})
	return accountsPasswords, nil
}
