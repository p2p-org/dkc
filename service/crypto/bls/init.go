package bls

import "github.com/herumi/bls-eth-go-binary/bls"

func init() {
	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)
}
