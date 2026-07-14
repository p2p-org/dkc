package bls

import (
	"github.com/herumi/bls-eth-go-binary/bls"
)

// Split represents key as a polynomial of degree threshold-1: the first
// coefficient is the key itself, the rest are random. Returned master keys
// are the serialized coefficients.
func Split(key []byte, threshold uint32) ([][]byte, [][]byte, error) {
	var sk bls.SecretKey

	if err := sk.Deserialize(key); err != nil {
		return nil, nil, err
	}

	masterPKs := [][]byte{sk.GetPublicKey().Serialize()}
	masterSKs := [][]byte{sk.Serialize()}

	for i := uint32(1); i < threshold; i++ {
		var coef bls.SecretKey
		coef.SetByCSPRNG() // Shouldn't be a zero (all keys will be equal in that case)
		masterSKs = append(masterSKs, coef.Serialize())
		masterPKs = append(masterPKs, coef.GetPublicKey().Serialize())
	}

	return masterSKs, masterPKs, nil
}

// Combine recovers the original key from at least threshold shards keyed by
// participant id
func Combine(shards map[uint64][]byte) ([]byte, error) {
	var subIDs []bls.ID
	var subSKs []bls.SecretKey
	for id, key := range shards {
		blsID, err := newBlsID(id)
		if err != nil {
			return nil, err
		}
		subIDs = append(subIDs, *blsID)

		var sk bls.SecretKey
		if err := sk.Deserialize(key); err != nil {
			return nil, err
		}
		subSKs = append(subSKs, sk)
	}

	var rk bls.SecretKey
	if err := rk.Recover(subSKs, subIDs); err != nil {
		return nil, err
	}

	return rk.Serialize(), nil
}

// SetupParticipants derives one key shard per participant id by evaluating
// the polynomial defined by masterSKs at that id
func SetupParticipants(masterSKs [][]byte, ids []uint64) (map[uint64][]byte, error) {
	mSKs := make([]bls.SecretKey, 0, len(masterSKs))
	for _, s := range masterSKs {
		var sk bls.SecretKey
		if err := sk.Deserialize(s); err != nil {
			return nil, err
		}
		mSKs = append(mSKs, sk)
	}

	shards := make(map[uint64][]byte, len(ids))
	for _, id := range ids {
		blsID, err := newBlsID(id)
		if err != nil {
			return nil, err
		}

		var sk bls.SecretKey
		if err := sk.Set(mSKs, blsID); err != nil {
			return nil, err
		}

		shards[id] = sk.Serialize()
	}

	return shards, nil
}
