# Dirk Key Converter (dkc)

[![Go Report Card](https://goreportcard.com/badge/github.com/p2p-org/dkc)](https://goreportcard.com/report/github.com/p2p-org/dkc)

Transform wallets between [distributed type](https://github.com/wealdtech/go-eth2-wallet-distributed) and [non-deterministic type](https://github.com/wealdtech/go-eth2-wallet-nd).

:bangbang: It is highly recommended to refrain from any operations on the validation keys and use the provided script only in critical situations to avoid any ponential risks of slashing.

## Table of Contents

- [Install](#install)
  - [Binaries](#binaries)
  - [Source](#source)
- [Usage](#usage)
  - [Config](#config)
  - [Combine](#combine)
  - [Split](#split)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

### Binaries

Binaries for the latest version of `dkc` can be obtained from [the releases page](https://github.com/p2p-org/dkc/releases/latest).

### Source

`dkc` can be built locally using the command

```sh
go build .
```
## Usage

:bangbang: Before you begin, make sure you backup keys and store recovered wallets and passwords securely.

`dkc` uses [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary). The pseudocode of using [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary) in `dkc` is available [here](./example/split.go) 

You can also test `dkc` on predefiened inputs [here](.github/workflows/func-tests.yml). 

### Config

An example config can be found [here](.github/examples/config.yaml)

```yaml
distributed-wallets:                #Distributed-wallets section
  path: ./wallet                    #Path to distributed wallet (Default: None)
  passphrases: ./passphrases.txt    #Path to file containing passphrases for unlocking/locking accounts (Default: None)
  threshold: 2                      #Threshlod value (Default: None)
  peers:                            #Peers dict, number of peers must be greater than threshold value (Default: None)
    10: old1:9091
    20: old2:9091
    30: old3:9091
nd-wallets:                         #Non-determenistic-wallets section                    
  path: ND_WALLETS                  #Path to non-determenistic wallet (Default: None)
  passphrases: ./passphrases.txt    #Path to file containing passphrases for unlocking/locking accounts (Default: None)
log-level: debug                    #Log-level (Default: INFO)
```

### Combine

Combine [distributed type](https://github.com/wealdtech/go-eth2-wallet-distributed) to [non-deterministic type](https://github.com/wealdtech/go-eth2-wallet-nd).

```sh
./dkc combine --config=config.yaml
```

### Split

Split [non-deterministic type](https://github.com/wealdtech/go-eth2-wallet-nd) to [distributed type](https://github.com/wealdtech/go-eth2-wallet-distributed)

```sh
./dkc split --config=config.yaml
```

## Maintainers

Sergey Yakovlev: [@selfuryon](https://github.com/selfuryon).

SpontaneousOverthrow: [@SpontaneousOverthrow](https://github.com/SpontaneousOverthrow).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/p2p-org/dkc/issues).

## License

[License](./LICENSE)
