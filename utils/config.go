package utils

import (
	"bytes"
	"os"

	"github.com/pkg/errors"
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
	if len(accountsPasswords) == 0 {
		err := errors.New(ErrorPassphrasesField)
		return nil, errors.Wrap(err, ErrorDWalletStructWrapper)
	}
	return accountsPasswords, nil
}

func (data *NDWalletConfig) Validate() error {
	if data.Path == "" {
		err := errors.New(ErrorPathField)
		return errors.Wrap(err, ErrorNDWalletStructWrapper)
	}

	if data.Passphrases == "" {
		err := errors.New(ErrorPassphrasesField)
		return errors.Wrap(err, ErrorNDWalletStructWrapper)
	}

	return nil
}

func (data *DWalletConfig) Validate() error {
	if data.Path == "" {
		err := errors.New(ErrorPathField)
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if data.Passphrases == "" {
		err := errors.New(ErrorPassphrasesField)
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(data.Peers) == 0 {
		err := errors.New(ErrorPeersField)
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if data.Threshold == 0 {
		err := errors.New(ErrorThresholdField)
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(data.Peers) < int(data.Threshold) {
		err := errors.New(ErrorNotEnoughPeers)
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	return nil
}
