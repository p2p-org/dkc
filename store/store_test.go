package store

import (
	"context"
	"fmt"
	"os"
	"testing"

	herumi "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/p2p-org/dkc/utils"
	"github.com/rs/zerolog"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
	"golang.org/x/sync/errgroup"
)

func TestMain(m *testing.M) {
	utils.Log = zerolog.Nop()
	if err := e2types.InitBLS(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// TestParallelSavePrivateKey guards the concurrency model of SavePrivateKey:
// all writers must share one wallet instance so that parallel imports cannot
// lose accounts index updates.
func TestParallelSavePrivateKey(t *testing.T) {
	const accounts = 8

	ctx := context.Background()
	s := &simpleStore{
		Type:        "non-deterministic",
		Label:       "ND Store",
		Path:        t.TempDir(),
		Passphrases: [][]byte{[]byte("test-passphrase")},
		Ctx:         ctx,
		cache:       NewWalletCache(),
	}

	if err := s.Create(); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.CreateWallet("W1"); err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}

	g := errgroup.Group{}
	for i := range accounts {
		g.Go(func() error {
			var sk herumi.SecretKey
			sk.SetByCSPRNG()
			return s.SavePrivateKey("W1", fmt.Sprintf("acc%d", i), sk.Serialize())
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("parallel SavePrivateKey: %v", err)
	}

	// Every account must be present both on disk and in the wallet index
	accs, wallets, err := getWalletsAccountsMap(ctx, s.Path)
	if err != nil {
		t.Fatalf("getWalletsAccountsMap: %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(wallets))
	}
	if len(accs) != accounts {
		t.Fatalf("expected %d accounts on disk, got %d", accounts, len(accs))
	}

	wallet, err := getWallet(s.Path, "W1")
	if err != nil {
		t.Fatalf("getWallet: %v", err)
	}
	byName := wallet.(types.WalletAccountByNameProvider)
	for i := range accounts {
		name := fmt.Sprintf("acc%d", i)
		if _, err := byName.AccountByName(ctx, name); err != nil {
			t.Errorf("account %s not found via wallet index: %v", name, err)
		}
	}
}

func TestGetAccountsPasswords(t *testing.T) {
	path := t.TempDir() + "/pass.txt"
	if err := os.WriteFile(path, []byte("pass1\npass2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	passphrases, err := getAccountsPasswords(path)
	if err != nil {
		t.Fatalf("getAccountsPasswords: %v", err)
	}
	// Trailing newline must not produce an empty passphrase
	if len(passphrases) != 2 {
		t.Fatalf("expected 2 passphrases, got %d", len(passphrases))
	}
	if string(passphrases[0]) != "pass1" || string(passphrases[1]) != "pass2" {
		t.Fatalf("unexpected passphrases: %q", passphrases)
	}

	if err := os.WriteFile(path, []byte("\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := getAccountsPasswords(path); err == nil {
		t.Fatal("expected error for file with only empty lines")
	}
}
