package utils

import (
	"github.com/pkg/errors"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

const (
	errorFailedToCreateWalletWrapper = "failed to create wallet"
)

type NDW struct {
	Type        string
	Passphrases [][]byte
}

type HDW struct {
	Type        string
	Passphrases [][]byte
}

type DW struct {
	Type        string
	Passphrases [][]byte
	Threshold   int
	Peers       map[uint64]string
}

type OutputWalletWrapper interface {
	Validate(wType string, wDist string) error
	types.WalletLocker
	Create(store types.Store, wType string, walletName string) (interface{}, error)
}

type InputWalletWrapper interface {
	Validate(wType string, wDist string) error
	types.WalletLocker
}

func (dw DW) ImportAccount() error {
	return nil
}

func (ndw NDW) ImportAccount() error {
	return nil
}

func (dw DW) Validate(wType string, wDist string) error {

	if len(dw.Passphrases) == 0 {
		err := ErrorPassphrasesField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(dw.Peers) == 0 {
		err := ErrorPeersField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if dw.Threshold == 0 {
		err := ErrorThresholdField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	if len(dw.Peers) < int(dw.Threshold) {
		err := ErrorNotEnoughPeers
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	return nil
}

func (hdw HDW) Validate(wType string, wDist string) error {

	if len(hdw.Passphrases) == 0 {
		err := ErrorPassphrasesField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	return nil
}

func (ndw NDW) Validate(wType string, wDist string) error {

	if len(ndw.Passphrases) == 0 {
		err := ErrorPassphrasesField
		return errors.Wrap(err, ErrorDWalletStructWrapper)
	}

	return nil
}

func (ndw NDW) Create(store types.Store, wType string, walletName string) (interface{}, error) {

	wallet, err := createWallet(store, "non-deterministic", walletName)
	if err != nil {
		return nil, err
	}

	return wallet.(types.WalletAccountImporter), err
}

func (dw DW) Create(store types.Store, wType string, walletName string) (interface{}, error) {

	wallet, err := createWallet(store, "distributed", walletName)
	if err != nil {
		return nil, err
	}

	return wallet.(types.WalletDistributedAccountImporter), err
}

func createWallet(store types.Store, wType string, walletName string) (types.Wallet, error) {
	err := e2wallet.UseStore(store)
	if err != nil {
		return nil, err
	}
	wallet, err := e2wallet.CreateWallet(walletName, e2wallet.WithType(wType))
	if err != nil {
		return nil, errors.Wrap(err, errorFailedToCreateWalletWrapper)
	}
	return wallet, nil
}
