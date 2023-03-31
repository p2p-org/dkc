# Dirk Key Converter (dkc)
Transform wallets between [distributed type](https://github.com/wealdtech/go-eth2-wallet-distributed) and [non-deterministic type](https://github.com/wealdtech/go-eth2-wallet-nd).

## Build
```
go build .
```
The output binary will be under the name: `./dkc`

## Usage
Before you begin, please use tmpfs for storing recovered wallets and passwords.
### Combine
- Prepare [config.yaml](./config.yaml).
    Update `peers` section(this data can be obtained from devops team)
    Update `distributed-wallets` this is the path where input wallets are located
    Update `passphrases` this is passwords list for the input and output wallets
    Update `nd-wallets` this is the path where output wallets will be located
- Run:
```
./dkc combine --config config.yaml
```
- Non-deterministic wallets will be located in the `nd-wallets` directory

### Split
- Prepare [config.yaml](./config.yaml).
    Update `peers` section(this data can be obtained from devops team)
    Update `nd-wallets` this is the path where input wallets will be located
    Update `passphrases` this is passwords list for the input and output wallets
    Update `distributed-wallets` this is the path where output wallets will be located
- Run:
```
./dkc combine --config config.yaml
```
- Distributed wallets will be located in the `distributed-wallets` directory

