package bls

import (
	"context"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/p2p-org/dkc/service"
)

func Split(ctx context.Context, key []byte, threshold uint32) (
	masterSKs [][]byte,
	masterPKs [][]byte,
) {
	var sk bls.SecretKey
	sk.Deserialize(key)
	masterPKs = append(masterPKs, sk.GetPublicKey().Serialize())
	masterSKs = append(masterSKs, sk.Serialize())

	for i := 1; i < int(threshold); i++ {
		var sk bls.SecretKey
		sk.SetByCSPRNG() // Shouldn't be a zero (all keys will be equal in that case)
		masterSKs = append(masterSKs, sk.Serialize())
		masterPKs = append(masterPKs, sk.GetPublicKey().Serialize())
	}

	return
}

func Sign(ctx context.Context, accounts []service.Account) []byte {
	var subSignatures []bls.Sign
	var subIDs []bls.ID
	var sig bls.Sign

	for _, account := range accounts {
		subIDs = append(
			subIDs,
			*newBlsID(account.ID),
		)

		var peerSig bls.Sign
		peerSig.Deserialize(account.Signature)
		subSignatures = append(
			subSignatures,
			peerSig,
		)
	}

	if err := sig.Recover(subSignatures, subIDs); err != nil {
		panic("rap")
	}

	return sig.Serialize()
}

func Recover(ctx context.Context, accounts []service.Account) ([]byte, error) {
	var subIDs []bls.ID
	var subSKs []bls.SecretKey
	for _, account := range accounts {
		subIDs = append(
			subIDs,
			*newBlsID(account.ID),
		)

		var mk bls.SecretKey
		mk.Deserialize(account.Key)

		subSKs = append(subSKs, mk)
	}

	var rk bls.SecretKey
	if err := rk.Recover(subSKs, subIDs); err != nil {
		panic(err)
	}

	return rk.Serialize(), nil
}
