package utils

const (
	ErrorPubKeyMatch          = "Pubkeys don't match"
	ErrorSignatureMatch       = "Signature doesn't match"
	ErrorNoPrivateKeyMsg      = "Account does not provide it's Private Key"
	ErrorAccountIsNotUnlocked = "Account is not unlocked"

	ErrorFailedToCreateWalletWrapper = "failed to create wallet"
	ErrorWalletDirWrapper            = "failed to get wallet dir"
	ErrorLoadStoreWrapper            = "can't load store"
	ErrorUseStoreWrapper             = "failed to use store"
	ErrorUnlockWrapper               = "failed to unlock account"
	ErrorImportWrapper               = "failed to import account"
	ErrorAccountSign                 = "failed to sign message"
	ErrorSignVerify                  = "failed to verify signature"
)
