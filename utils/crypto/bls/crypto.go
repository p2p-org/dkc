package bls

import (
	"context"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/p2p-org/dkc/utils"
)

func Split(ctx context.Context, key []byte, threshold uint32) ([][]byte, [][]byte, error) {
	var sk bls.SecretKey

	err := sk.Deserialize(key)
	if err != nil {
		return nil, nil, err
	}

	masterPKs := append([][]byte{}, sk.GetPublicKey().Serialize())
	masterSKs := append([][]byte{}, sk.Serialize())

	for i := 1; i < int(threshold); i++ {
		var sk bls.SecretKey
		sk.SetByCSPRNG() // Shouldn't be a zero (all keys will be equal in that case)
		masterSKs = append(masterSKs, sk.Serialize())
		masterPKs = append(masterPKs, sk.GetPublicKey().Serialize())
	}

	return masterSKs, masterPKs, nil
}

func Sign(ctx context.Context, accounts []utils.Account) ([]byte, error) {
	var subSignatures []bls.Sign
	var subIDs []bls.ID
	var sig bls.Sign

	for _, account := range accounts {
		blsID, err := newBlsID(account.ID)
		if err != nil {
			return nil, err
		}
		subIDs = append(
			subIDs,
			*blsID,
		)

		var peerSig bls.Sign
		err = peerSig.Deserialize(account.Signature)
		if err != nil {
			return nil, err
		}
		subSignatures = append(
			subSignatures,
			peerSig,
		)
	}

	if err := sig.Recover(subSignatures, subIDs); err != nil {
		return nil, err
	}

	return sig.Serialize(), nil
}

func Recover(ctx context.Context, accounts []utils.Account) ([]byte, error) {
	var subIDs []bls.ID
	var subSKs []bls.SecretKey
	for _, account := range accounts {
		blsID, err := newBlsID(account.ID)
		if err != nil {
			return nil, err
		}
		subIDs = append(
			subIDs,
			*blsID,
		)

		var mk bls.SecretKey
		err = mk.Deserialize(account.Key)
		if err != nil {
			return nil, err
		}

		subSKs = append(subSKs, mk)
	}

	var rk bls.SecretKey
	if err := rk.Recover(subSKs, subIDs); err != nil {
		return nil, err
	}

	return rk.Serialize(), nil
}

func SetupParticipants(masterSKs [][]byte, masterPKs [][]byte, ids []uint64, threshold int) (
	[]utils.Account, error,
) {
	var mSKs []bls.SecretKey
	var accounts = []utils.Account{}

	for _, s := range masterSKs {
		var sk bls.SecretKey
		err := sk.Deserialize(s)
		if err != nil {
			return nil, err
		}
		mSKs = append(mSKs, sk)
	}

	for i := 0; i < threshold; i++ {
		id, err := newBlsID(ids[i])
		if err != nil {
			return nil, err
		}

		var sk bls.SecretKey
		if err := sk.Set(mSKs, id); err != nil {
			return nil, err
		}

		accounts = append(accounts,
			utils.Account{
				Key:       sk.Serialize(),
				ID:        ids[i],
				Signature: nil,
			})
	}

	return accounts, nil
}
