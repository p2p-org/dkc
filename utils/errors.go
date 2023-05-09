package utils

const (
	ErrorPubKeyMatch          = "pubkeys don't match"
	ErrorSignatureMatch       = "signature doesn't match"
	ErrorNoPrivateKeyMsg      = "account does not provide it's private key"
	ErrorAccountIsNotUnlocked = "account is not unlocked"
	ErrorPassphrasesField     = "passphrases field is empty"
	ErrorPathField            = "path field is empty"
	ErrorPeersField           = "peers field is empty"
	ErrorThresholdField       = "threshold field is empty"
	ErrorPassphraseEmpty      = "passphrase is empty"
	ErrorNotEnoughPeers       = "current peer value is not enough for threshold value"

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
