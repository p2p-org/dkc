package bls

import (
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func newBlsID(id uint64) *bls.ID {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	err := res.SetLittleEndian(buf[:])
	if err != nil {
		panic(err)
	}
	return &res
}

//func setupParticipants(masterSKs []bls.SecretKey, masterPKs []bls.PublicKey, ids []uint64, threshold int) (
//	participantsIDs []bls.ID,
//	participantsSKs []bls.SecretKey,
//	participantsPKs []bls.PublicKey,
//) {
//	for i := 0; i < threshold; i++ {
//		id := newBlsID(uint64(ids[i]))
//
//		participantsIDs = append(participantsIDs, *id)
//		var sk bls.SecretKey
//		if err := sk.Set(masterSKs, id); err != nil {
//			log.Fatalf("Failed to Set secret key: %s", err)
//		}
//		participantsSKs = append(participantsSKs, sk)
//
//		var pk bls.PublicKey
//		if err := pk.Set(masterPKs, id); err != nil {
//			log.Fatalf("Failed to Set public key: %s", err)
//		}
//		participantsPKs = append(participantsPKs, pk)
//	}
//
//	return
//}

func setupMasterKeys(masterSK []byte, threshold uint32) (masterSKs [][]byte, masterPKs [][]byte) {
	var sk bls.SecretKey
	sk.Deserialize(masterSK)
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
