package utils

import (
	"context"

	"github.com/pkg/errors"

	types "github.com/wealdtech/go-eth2-wallet-types/v2"
)

const (
	signingString = "bkeCE2vRuTxxc5RpzrvLzoU5EgulV7uk3zMnt5MP9MgsXBaif9mUQcf7rZGC5mNj9lBqQ2s"
)

func CreateNDAccount(
	wallet NDWallet,
	name string,
	masterSK []byte,
	passphrase []byte,
) (types.Account, error) {

	err := wallet.Unlock(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, ErrorUnlockWrapper)
	}

	defer func() {
		err = wallet.Lock(context.Background())
	}()

	account, err := wallet.ImportAccount(context.Background(),
		name,
		masterSK,
		passphrase,
	)
	if err != nil {
		return nil, errors.Wrap(err, ErrorImportWrapper)
	}

	return account, nil
}

func CreateDAccount(
	wallet DWallet,
	name string,
	masterPKs [][]byte,
	masterSK []byte,
	threshold uint32,
	peers map[uint64]Peer,
	passphrase string,
) (types.Account, error) {

	err := wallet.Unlock(context.Background(), nil)
	if err != nil {
		return nil, errors.Wrap(err, ErrorUnlockWrapper)
	}

	defer func() {
		err = wallet.(types.WalletLocker).Lock(context.Background())
	}()

        peerMap := make(map[uint64]string, 0)
        for id, peer := range peers {
                peerMap[id] = peer.Host
        }

        passBytes := []byte(passphrase)

	account, err := wallet.ImportDistributedAccount(context.Background(),
		name,
		masterSK,
		threshold,
		masterPKs,
		peerMap,
		passBytes)
	if err != nil {
		return nil, errors.Wrap(err, ErrorImportWrapper)
	}

	return account, nil
}

func AccountSign(ctx context.Context, acc types.Account, passphrases [][]byte) ([]byte, error) {
	account, err := unlockAccount(ctx, acc, passphrases)
	if err != nil {
		return nil, errors.Wrap(err, ErrorUnlockWrapper)
	}

	accountSigner := account.(types.AccountSigner)
	signedData, err := accountSigner.Sign(ctx, []byte(signingString))
	if err != nil {
		return nil, errors.Wrap(err, ErrorAccountSign)
	}

	if !signedData.Verify([]byte(signingString), acc.PublicKey()) {
		return nil, errors.Wrap(err, ErrorSignVerify)
	}

	return signedData.Marshal(), nil
}

func GetAccountKey(ctx context.Context, account types.Account, passphrases [][]byte) ([]byte, error) {
	privateKeyProvider, isPrivateKeyProvider := account.(types.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		err := ErrorNoPrivateKeyMsg
		return nil, err
	}

	_, err := unlockAccount(ctx, account, passphrases)
	if err != nil {
		return nil, err
	}

	key, err := privateKeyProvider.PrivateKey(ctx)
	if err != nil {
		return nil, err
	}

	return key.Marshal(), nil
}

func GetAccountCompositePubkey(account types.Account) ([]byte, error) {
	compositePublicKeyProvider, isCompositePublicKeyProvider := account.(types.AccountCompositePublicKeyProvider)
	if !isCompositePublicKeyProvider {
		err := ErrorNoPrivateKeyMsg
		return nil, err
	}

	pubKey := compositePublicKeyProvider.CompositePublicKey()

	return pubKey.Marshal(), nil
}

func GetAccountPubkey(account types.Account) (data []byte, err error) {
	data = account.PublicKey().Marshal()
	return
}

func unlockAccount(ctx context.Context, acc types.Account, passphrases [][]byte) (types.Account, error) {
	account := acc
	if locker, isLocker := account.(types.AccountLocker); isLocker {
		unlocked, err := locker.IsUnlocked(ctx)
		if err != nil {
			return nil, err
		}
		if !unlocked {
			for _, passphrase := range passphrases {
				err = locker.Unlock(ctx, passphrase)
				if err == nil {
					unlocked = true
					break
				}
			}
			if !unlocked {
				err = ErrorAccountIsNotUnlocked
				return nil, err
			}
		}
	}

	return account, nil
}
