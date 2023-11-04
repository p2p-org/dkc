package utils

import (
	"bytes"
	"context"
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
	WalletName  string
}

type W struct {
	Path        string
	Type        string
	Passphrases string
}

type ConvertRuntime struct {
	Ctx     context.Context
	InputW  W
	OutputW W
}

func (W *W) wValidate() error {
	if W.Path == "" {
		return ErrorPathField
	}

	if W.Type == "" {
		return ErrorEmptyType
	}

	if len(W.Passphrases) == 0 {
		return ErrorPassphraseEmpty
	}

	return nil

}

func (CR *ConvertRuntime) Validate() error {
	LogConvert.Debug().Msg("validating input wallet")
	CR.InputW.wValidate()
	if CR.InputW.Passphrases == "" {
		return ErrorInputPassphrasesIsEmpty
	}

	LogConvert.Debug().Msg("validating output wallet")
	CR.OutputW.wValidate()
	if CR.OutputW.Passphrases == "" {
		return ErrorOutputPassphrasesIsEmpty
	}

	LogConvert.Debug().Msg("validating types for both wallets")
	if CR.InputW.Type != CR.OutputW.Type {
		return ErrorSameWalletType
	}

	LogConvert.Debug().Msg("validating dirs for both wallets")
	if CR.OutputW.Path != CR.OutputW.Path {
		return ErrorSameDirs
	}

	return nil
}
func GetAccountsPasswords(path string) ([][]byte, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	accountsPasswords := bytes.Split(content, []byte{'\n'})
	if len(accountsPasswords) == 0 {
		err := ErrorPassphrasesField
		return nil, errors.Wrap(err, ErrorDWalletStructWrapper)
	}
	return accountsPasswords, nil
}

func (data *NDWalletConfig) Validate() error {
	if data.Path == "" {
		err := ErrorPathField
		return errors.Wrap(err, ErrorNDWalletStructWrapper)
	}

	if data.Passphrases == "" {
		err := ErrorPassphrasesField
		return errors.Wrap(err, ErrorNDWalletStructWrapper)
	}

	return nil
}

func (data *DWalletConfig) Validate() error {
	if data.Path == "" {
		err := ErrorPathField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if data.Passphrases == "" {
		err := ErrorPassphrasesField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(data.Peers) == 0 {
		err := ErrorPeersField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if data.Threshold == 0 {
		err := ErrorThresholdField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(data.Peers) < int(data.Threshold) {
		err := ErrorNotEnoughPeers
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	return nil
}
