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

func newBlsID(id uint64) *bls.ID {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	if err := res.SetLittleEndian(buf[:]); err != nil {
		panic(err)
	}
	return &res
}

func sample() {
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
	sample()
}
