// nolint
package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/herumi/bls-eth-go-binary/bls"
	// e2wallet "github.com/wealdtech/go-eth2-wallet"
	// distributed "github.com/wealdtech/go-eth2-wallet-distributed"
	// keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	// filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	// e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

const THRESHOLD = 2
const PARTICIPANTS = 3
const KEYSTR = "3eb84bbe03db1c6341c490142a647655f33983ed693d0f43c696ed0378fdc492"

func sample1() {
	msg := []byte("Hello")
	var origin_key bls.SecretKey
	//origin_key.SetByCSPRNG()
	origin_key_byte, _ := hex.DecodeString(KEYSTR)
	origin_key.Deserialize(origin_key_byte)
	origin_pub := origin_key.GetPublicKey()

	fmt.Printf("origin key=%v\n", origin_key.SerializeToHexStr())
	fmt.Printf("origin pub=%v\n", origin_pub.SerializeToHexStr())

	masterSKs, masterPKs := setupMasterKeys(origin_key)

	partIDs, partSKs, partPKs, partSigns := setupParticipants(masterSKs, masterPKs, msg)
	// Check
	if err := checkKeys(origin_key, msg, partIDs, partSKs, partPKs, partSigns); err != nil {
		log.Fatalf("failed to checkKeys: %s", err)
	}
	log.Println("keys check success")
}

func setupParticipants(masterSKs []bls.SecretKey, masterPKs []bls.PublicKey, msg []byte) (
	participantsIDs []bls.ID,
	participantsSKs []bls.SecretKey,
	participantsPKs []bls.PublicKey,
	signatures []bls.Sign,
) {
	for i := 0; i < PARTICIPANTS; i++ {
		id := newBlsID(uint64(i + 1))

		participantsIDs = append(participantsIDs, *id)
		var sk bls.SecretKey
		if err := sk.Set(masterSKs, id); err != nil {
			log.Fatalf("Failed to Set secret key: %s", err)
		}
		participantsSKs = append(participantsSKs, sk)
		fmt.Printf("partsk[%d]=%v\n", i, sk.SerializeToHexStr())

		var pk bls.PublicKey
		if err := pk.Set(masterPKs, id); err != nil {
			log.Fatalf("Failed to Set public key: %s", err)
		}
		participantsPKs = append(participantsPKs, pk)
		fmt.Printf("partpk[%d]=%v\n", i, pk.SerializeToHexStr())

		sig := sk.SignByte(msg)
		signatures = append(signatures, *sig)
	}

	return
}

func setupMasterKeys(masterSK bls.SecretKey) (masterSKs []bls.SecretKey, masterPKs []bls.PublicKey) {
	masterSKs = append(masterSKs, masterSK)

	for i := 1; i < THRESHOLD; i++ {
		var sk bls.SecretKey
		sk.SetByCSPRNG() // Shouldn't be a zero (all keys will be equal in that case)
		masterSKs = append(masterSKs, sk)
		fmt.Printf("mk[%d]=%v\n", i, sk.SerializeToHexStr())
	}

	masterPKs = bls.GetMasterPublicKey(masterSKs)

	return
}

func checkKeys(
	masterSK bls.SecretKey,
	msg []byte,
	participantsIDs []bls.ID,
	participantsSKs []bls.SecretKey,
	participantsPKs []bls.PublicKey,
	signatures []bls.Sign,
) error {
	indexPairs := [][]uint32{{1, 2}, {0, 2}, {0, 1}}
	// indexPairs := [][]uint32{
	// {0, 1}, {0, 2}, {0, 3}, {0, 4},
	// {1, 2}, {1, 3}, {1, 4},
	// {2, 3}, {3, 4},
	// }
	for idx, indexPair := range indexPairs {
		var (
			subIDs  []bls.ID
			subSKs  []bls.SecretKey
			subPKs  []bls.PublicKey
			subSigs []bls.Sign
		)

		for i := 0; i < 2; i++ {
			idx := indexPair[i]
			subIDs = append(subIDs, participantsIDs[idx])
			subSKs = append(subSKs, participantsSKs[idx])
			subPKs = append(subPKs, participantsPKs[idx])
			subSigs = append(subSigs, signatures[idx])
		}

		var sec bls.SecretKey
		var pub bls.PublicKey
		var sig bls.Sign

		if err := sec.Recover(subSKs, subIDs); err != nil {
			return fmt.Errorf("failed to Recover priv: %w", err)
		}

		if err := pub.Recover(subPKs, subIDs); err != nil {
			return fmt.Errorf("failed to Recover pub: %w", err)
		}

		if err := sig.Recover(subSigs, subIDs); err != nil {
			return fmt.Errorf("failed to Recover signature: %w", err)
		}

		if !sig.VerifyByte(masterSK.GetPublicKey(), msg) {
			return fmt.Errorf("failed to verify signature for index pair %d", idx)
		}

		fmt.Printf("------\n")
		fmt.Printf("%d: mk=%v\n", idx, masterSK.SerializeToHexStr())
		fmt.Printf("%d: rk=%v\n", idx, sec.SerializeToHexStr())
		fmt.Printf("%d: mp=%v\n", idx, masterSK.GetPublicKey().SerializeToHexStr())
		fmt.Printf("%d: rp=%v\n", idx, pub.SerializeToHexStr())
	}

	return nil
}

func newBlsID(id uint64) *bls.ID {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	if err := res.SetLittleEndian(buf[:]); err != nil {
		panic(err)
	}
	return &res
}

func sample2() {
	//	msg := []byte("Hello")
	var mk1 bls.SecretKey
	var mk2 bls.SecretKey
	mk1_byte, _ := hex.DecodeString("3eb84bbe03db1c6341c490142a647655f33983ed693d0f43c696ed0378fdc492")
	mk2_byte, _ := hex.DecodeString("56826b2549ba1c26eb4dcbb73807fc81d49d8c754c4a034a578bd808b0d2f56c")
	mk1.Deserialize(mk1_byte)
	mk2.Deserialize(mk2_byte)

	fmt.Printf("mk1=%v\nmk2=%v\n", mk1.SerializeToHexStr(), mk2.SerializeToHexStr())

	masterSKs := []bls.SecretKey{mk1, mk2}
	var partSKs []bls.SecretKey
	// Generate
	for i := 0; i < 3; i++ {
		id := newBlsID(uint64(i + 1))
		var sk bls.SecretKey
		if err := sk.Set(masterSKs, id); err != nil {
			log.Fatalf("Failed to Set secret key: %s", err)
		}
		partSKs = append(partSKs, sk)
		fmt.Printf("partsk[%d]=%v\n", i, sk.SerializeToHexStr())
	}
	//Recover

	subSKs := []bls.SecretKey{partSKs[0], partSKs[1]}
	subIDs := []bls.ID{*newBlsID(1), *newBlsID(2)}
	var rk bls.SecretKey
	if err := rk.Recover(subSKs, subIDs); err != nil {
	}
	fmt.Printf("Recovered key=%v\n", rk.SerializeToHexStr())

}

func main() {
	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)
	//sample1()
	sample2()
}
