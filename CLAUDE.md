# dkc - Dirk Key Converter

CLI that converts Ethereum validator wallets between types: distributed
(dirk threshold shards), non-deterministic, keystore, and hierarchical
deterministic (input only). Distributed conversion splits/combines BLS keys
via Shamir-style sharing (herumi/bls-eth-go-binary, cgo).

## Build and test

- Dev shell: `nix develop` (Go, gcc, ethdo). CGO is required.
- Build: `go build .`; nix package: `nix build .#dkc`.
- **The nix flake builds from the git index**: new files must be `git add`ed
  or `nix build` fails with "undefined" errors while `go build` works.
- Unit tests: `go test ./...` (BLS roundtrip, parallel wallet imports).
- Functional tests: configs in `.github/examples/` - run
  `dkc convert --config <pair>.yaml` from that directory; verify by comparing
  ethdo signatures between input and output wallets (see func-tests.yaml).
  Delete generated output dirs afterwards.
- Vendor hash: after changing go.mod, recompute `vendorHash` in
  `nix/package.nix` (break it, build, copy the "got:" value).

## Architecture

- `cmd/` - cobra commands; `cmd/convert/process.go` is the conversion
  pipeline: parallel per-account GetPrivateKey -> SavePrivateKey, bounded by
  `errgroup.SetLimit(runtime.NumCPU())`.
- `store/` - `ComposedStore` wraps one or more `AtomicStore`s:
  - `simpleStore` (simple.go) - nd and keystore wallets (full keys);
  - `hdStore` (hd.go) - input only, keys derive from mnemonic;
  - `DistributedAtomicStore` (distributed.go) - one per dirk peer; the
    composed store splits keys into shards (crypto/bls) and saves one shard
    per peer.
  - `WalletCache` (cache.go) - populated for input stores at construction
    (accounts pre-unlocked in a worker pool); caches open wallet instances.
- `crypto/bls/` - Split/SetupParticipants/Combine; participant ids are the
  peer ids from config and must be non-zero.
- Config is viper YAML (see README); passphrase files are newline-separated,
  empty lines ignored.

## Critical invariants (learned from production races)

- Never use `e2wallet.UseStore` (global mutable store, racy). Always pass
  the store explicitly: `e2wallet.OpenWallet(name, e2wallet.WithStore(s))`.
- All writers to a wallet must share one instance via
  `WalletCache.GetOrOpenWallet`/`PutWallet`: parallel imports serialize on
  the wallet's own mutex; separate instances lose accounts-index updates
  (read-modify-write on the index file).
- Iterate config peer maps in sorted-id order - the first atomic store is
  the source for GetPath/GetAccounts of the composed store.
- The wealdtech filesystem store has no inter-process locking and writes
  files non-atomically: never run two dkc processes over the same store dir.
- Wallets/accounts are unlocked once and never re-locked during a run;
  re-locking breaks subsequent imports.
