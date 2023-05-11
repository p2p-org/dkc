package utils

import "github.com/pkg/errors"

var (
	ErrorPubKeyMatch          = errors.New("pubkeys don't match")
	ErrorSignatureMatch       = errors.New("signature doesn't match")
	ErrorNoPrivateKeyMsg      = errors.New("account does not provide it's private key")
	ErrorAccountIsNotUnlocked = errors.New("account is not unlocked")
	ErrorPassphrasesField     = errors.New("passphrases field is empty")
	ErrorPathField            = errors.New("path field is empty")
	ErrorPeersField           = errors.New("peers field is empty")
	ErrorThresholdField       = errors.New("threshold field is empty")
	ErrorPassphraseEmpty      = errors.New("passphrase is empty")
	ErrorNotEnoughPeers       = errors.New("current peer value is not enough for threshold value")
	ErrorSameDirs             = errors.New("same dir for d-wallets and nd-wallets")
)

const (
	ErrorNDWalletStructWrapper       = "failed to validate NDWalletStruct"
	ErrorDWalletStructWrapper        = "failed to validate DWalletStruct"
	ErrorFailedToCreateWalletWrapper = "failed to create wallet"
	ErrorWalletDirWrapper            = "failed to get wallet dir"
	ErrorLoadStoreWrapper            = "can't load store"
	ErrorUseStoreWrapper             = "failed to use store"
	ErrorUnlockWrapper               = "failed to unlock account"
	ErrorImportWrapper               = "failed to import account"
	ErrorAccountSign                 = "failed to sign message"
	ErrorSignVerify                  = "failed to verify signature"
)
