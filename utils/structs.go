package utils

import (
	"context"
)

type W struct {
	Path        string
	Type        string
	Passphrases [][]byte
}

type ConvertRuntime struct {
	Ctx     context.Context
	InputW  W
	OutputW W
}

func (W *W) Validate() error {
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

func (CR *ConvertRuntime) validate() error {
	if CR.InputW.Type != CR.OutputW.Type {
		return ErrorSameWalletType
	}

	if CR.OutputW.Path != CR.OutputW.Path {
		return ErrorSameDirs
	}

	return nil
}
