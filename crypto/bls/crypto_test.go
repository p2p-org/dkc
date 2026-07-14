package bls

import (
	"bytes"
	"os"
	"testing"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func TestMain(m *testing.M) {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestSplitCombineRoundtrip(t *testing.T) {
	var sk bls.SecretKey
	sk.SetByCSPRNG()
	original := sk.Serialize()

	const threshold = 2
	ids := []uint64{10, 20, 30}

	masterSKs, masterPKs, err := Split(original, threshold)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(masterSKs) != threshold || len(masterPKs) != threshold {
		t.Fatalf("expected %d master keys, got %d SKs / %d PKs", threshold, len(masterSKs), len(masterPKs))
	}

	shards, err := SetupParticipants(masterSKs, ids)
	if err != nil {
		t.Fatalf("SetupParticipants: %v", err)
	}
	if len(shards) != len(ids) {
		t.Fatalf("expected %d shards, got %d", len(ids), len(shards))
	}

	// Any subset of >= threshold shards must recover the original key
	subsets := [][]uint64{{10, 20}, {20, 30}, {10, 30}, {10, 20, 30}}
	for _, subset := range subsets {
		sub := map[uint64][]byte{}
		for _, id := range subset {
			sub[id] = shards[id]
		}
		combined, err := Combine(sub)
		if err != nil {
			t.Fatalf("Combine(%v): %v", subset, err)
		}
		if !bytes.Equal(combined, original) {
			t.Fatalf("Combine(%v) did not recover the original key", subset)
		}
	}

	// A single shard must NOT recover the original key
	combined, err := Combine(map[uint64][]byte{10: shards[10]})
	if err != nil {
		t.Fatalf("Combine(single): %v", err)
	}
	if bytes.Equal(combined, original) {
		t.Fatal("a single shard recovered the original key: threshold not enforced")
	}
}

func TestSplitRejectsInvalidKey(t *testing.T) {
	if _, _, err := Split([]byte("not a key"), 2); err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestZeroParticipantIDRejected(t *testing.T) {
	var sk bls.SecretKey
	sk.SetByCSPRNG()

	masterSKs, _, err := Split(sk.Serialize(), 2)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}

	// id 0 evaluates the sharing polynomial at x=0, which is the key itself
	if _, err := SetupParticipants(masterSKs, []uint64{0, 10}); err == nil {
		t.Fatal("expected error for zero participant id in SetupParticipants")
	}
	if _, err := Combine(map[uint64][]byte{0: masterSKs[0]}); err == nil {
		t.Fatal("expected error for zero participant id in Combine")
	}
}
