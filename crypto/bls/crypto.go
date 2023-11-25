package bls

import (
	"context"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func Split(ctx context.Context, key []byte, threshold uint32) ([][]byte, [][]byte, error) {
	var sk bls.SecretKey

	err := sk.Deserialize(key)
	if err != nil {
		return nil, nil, err
	}

	masterPKs := append([][]byte{}, sk.GetPublicKey().Serialize())
	masterSKs := append([][]byte{}, sk.Serialize())

	// We assume that peers
	for i := 1; i < int(threshold); i++ {
		var sk bls.SecretKey
		sk.SetByCSPRNG() // Shouldn't be a zero (all keys will be equal in that case)
		masterSKs = append(masterSKs, sk.Serialize())
		masterPKs = append(masterPKs, sk.GetPublicKey().Serialize())
	}

	return masterSKs, masterPKs, nil
}

func Combine(ctx context.Context, accounts map[uint64][]byte) ([]byte, error) {
	var subIDs []bls.ID
	var subSKs []bls.SecretKey
	for id, key := range accounts {
		blsID, err := newBlsID(id)
		if err != nil {
			return nil, err
		}
		subIDs = append(
			subIDs,
			*blsID,
		)

		var mk bls.SecretKey
		err = mk.Deserialize(key)
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
	map[uint64][]byte, error,
) {
	var mSKs []bls.SecretKey
	peersIDs := map[uint64][]byte{}

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

		peersIDs[ids[i]] = sk.Serialize()
	}

	return peersIDs, nil
}
