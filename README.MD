# dkc
Dirk Key Converter

## Build
```
go build .
```
The output binary will be under the name: `./dkc`

## Usage
### Combine
- Mount `tmpfs` file system on `./wallets`. This is the path where input `distributed wallets` will be located.
- Mount `tmpfs` file system on `./restoredwallets`. This is the path where output `non-deterministic wallets` will be located.
- Prepare [config.yaml](./config.yaml). Update `peers:` section(this data can be obtained from devops team)
- Place password file in `./wallets/passphrases.txt`
- For each `peer` create a directory inside the [./wallets](./wallets) directory. The name of the directory must contain only the `peer` name without the port number.
- Populate `wallets/peer` directories with data
- Run `./dkc combine`
- Combined wallets will be located in the `./restoredwallets` directory
- Additional:
    Check number the of accounts in each wallet, it must equal the number of accounts for each peer's wallet.
    Check `composite pubkey(first element of verificationvector json field)` for each peer account with `pubkey` field in combined accounts

### Split
- Mount `tmpfs` file system on `./restoredwallets`. This is the path where input `non-deterministic wallets` will be located.
- Mount `tmpfs` file system on `./newwallets`. This is the path where output `distributed wallets` will be located.
- Prepare [config.yaml](./config.yaml). Update `peers:` section(this data can be obtained from devops team)
- Place password file in `./wallets/passphrases.txt`
- Populate [./restoredwallets](./restoredwallets) directory
- Run `./dkc split`
- Combined wallets will be located in the `./newwallets` directory
- Additional:
    Check the number of accounts in each peer's wallet, it must equal to the number of accounts for each `non-deterministic wallet`.
    Check `pubkey` for the input accounts with the `composite pubkey(first element of verificationvector json field)` of outputs accounts for each peer.

