package utils

import (
	"bytes"
	"context"
	"os"

	"github.com/pkg/errors"
)

type ConvertRuntime struct {
	Ctx     context.Context
	InputW  InputWalletWrapper
	OutputW OutputWalletWrapper
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
