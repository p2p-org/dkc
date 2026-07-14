# Dirk Key Converter (dkc)

[![Go Report Card](https://goreportcard.com/badge/github.com/p2p-org/dkc)](https://goreportcard.com/report/github.com/p2p-org/dkc)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/p2p-org/dkc/ci.yaml?label=ci)

A command-line tool for converting Ethereum wallets from:
- [distributed](https://github.com/wealdtech/go-eth2-wallet-distributed)
- [non-deterministic](https://github.com/wealdtech/go-eth2-wallet-nd)
- [hierarchical deterministic](https://github.com/wealdtech/go-eth2-wallet-hd)
- [keystore](https://github.com/wealdtech/go-eth2-wallet-keystore)

to:

- [distributed](https://github.com/wealdtech/go-eth2-wallet-distributed)
- [non-deterministic](https://github.com/wealdtech/go-eth2-wallet-nd)
- [keystore](https://github.com/wealdtech/go-eth2-wallet-keystore)

> [!CAUTION]
> It is highly recommended to refrain from any operations on the validation keys and use the provided script only in critical situations to avoid any potential risks of slashing.

## Table of Contents

- [Install](#install)
  - [Binaries](#binaries)
  - [Source](#source)
- [Usage](#usage)
  - [Config](#config)
  - [File Structure](#file-structure)
  - [Convert](#convert)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

### Binaries

Binaries for the latest version of `dkc` can be obtained from [the releases page](https://github.com/p2p-org/dkc/releases/latest).

### Nix

If you have nix flakes enabled you can run `dkc` using this command:

```sh
$ nix run github:p2p-org/dkc
```

### Source

`dkc` can be built locally using the command

```sh
$ go build .
```

## Usage

> [!CAUTION]
> Before you begin, make sure you backup keys and store recovered wallets and passwords securely.

`dkc` uses [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary). You can test `dkc` on predefined inputs [here](.github/workflows/func-tests.yaml).

### Config

An example config for each pair could be found [here](.github/examples/)

The example below shows how to convert wallets from `distributed` type to `non-deterministic` type (note: `hierarchical deterministic` can only be used as input - HD keys are derived from a mnemonic and cannot be imported).

`base-dir` for `distributed` wallets is located in `./i_wallet` see more info about `distributed` wallet file structure [here](#distributed)

`base-dir` for `non-deterministic` wallets is located in `./o_wallet` see more info about `non-deterministic` wallet file structure [here](#non-deterministic)

```yaml
input:                              #Input section
  store:                            #Store section
    path: ./i_wallet                #Location of input wallets
  wallet:                           #Wallet section
    type: distributed               #Type for input wallet. Valid types are: distributed, non-deterministic, hierarchical deterministic, keystore
    threshold: 2                    #Threshold number. It must be len(peers)/2 < threshold < len(peers) 
    peers:                          #Peers section
      10:                           #Peer ID is used for generating bls participants
        name: old1:9091             #Peer name is used for generating bls participants. All wallets for this peer are located in ./i_wallet/old1
        passphrases:                #Passphrases section
         path: ./peer1.txt          #Path to passphrases file(you can use separate passphrases file for each peer)
         index: 1                   #Password index in passphrases file. If not provided will use all passwords for unlocking wallets and only first password for creating accounts (Default: Using all passwords provided in passphrases file)
      20: 
        name: old2:9091
        passphrases: 
         path: ./peer2.txt
         index: 1
      30:
        name: old3:9091
        passphrases: 
         path: ./peer3.txt
         index: 1
output:                             #Output Wallet Section
  store:
    path: ./o_wallet
  wallet:
    type: non-deterministic         #Valid output types are: distributed, non-deterministic, keystore
    passphrases:
      path: ./o_passphrases.txt
log-level: debug                    #Log-level (Default: INFO)
```

### File Structure

#### Distributed

This wallet type can be used as input or output wallet. More information about this wallet type is provided [here](https://github.com/wealdtech/go-eth2-wallet-distributed)

The following is an [example](.github/examples/distributed-to-nd/distributed) file structure if one were to combine threshold keys. `test1`, `test2`, `test3` are each the `base-dir` wallet directory from the dirk instance. `test1`, `test2`, `test3` are each their own wallet directory. 

`ethdo wallet list --base-dir ./distributed-to-nd/distributed` should return nothing.

The subdirectories of `wallet` folder are the actual `ethdo` wallets: 

```
$ ethdo wallet list --base-dir distributed-to-nd/distributed/test1
Wallet2
Wallet3
Wallet1
$ ethdo wallet list --base-dir distributed-to-nd/distributed/test2
Wallet2
Wallet3
Wallet1
$ ethdo wallet list --base-dir distributed-to-nd/distributed/test3
Wallet1
Wallet2
Wallet3
```

*Importantly, the names of each wallet folder must correspond with the values in `*.wallet.peers` defined in the config.*

The keys within each wallet must also have the same name

All of the keys with corresponding names (ex: `name = 1`) should be the threshold keys corresponding to the same composite public key. 

#### Hierarchical Deterministic

This wallet type can be used only as input wallet. More information about this wallet type is provided [here](https://github.com/wealdtech/go-eth2-wallet-hd)

The following is an [example](.github/examples/hd-to-distributed/hd) file structure. The `base-dir` wallet directory is `hd-to-distributed/hd` 

```
$ ethdo wallet --base-dir hd-to-distributed/hd list
Wallet1
Wallet3
Wallet2
```


#### Non-Deterministic

This wallet type can be used as input or output wallet. More information about this wallet type is provided [here](https://github.com/wealdtech/go-eth2-wallet-nd)

The following is an [example](.github/examples/nd-to-distributed/nd) file structure. The `base-dir` wallet directory is `nd-to-distributed/nd`

```
$ ethdo wallet --base-dir nd-to-distributed/nd list
Wallet1
Wallet3
Wallet2
```

#### Keystore

This wallet type can be used as input or output wallet. More information about this wallet type is provided [here](https://github.com/wealdtech/go-eth2-wallet-keystore)

Keystore wallets store individual keys non-deterministically in keystore format. The file structure is similar to non-deterministic wallets:

```
$ ethdo wallet --base-dir keystore-to-distributed/keystore list
Wallet1
Wallet2
Wallet3
```

Configuration for keystore wallets:
```yaml
input:
  store:
    path: ./keystore_wallet
  wallet:
    type: keystore
    passphrases:
      path: ./keystore_passwords.txt
      index: 0  # Optional: use specific password index
```

### Convert

After preparing config and backing up keys simply run:

```sh
./dkc convert --config=config.yaml
```

Useful flags:

- `--log-level` - overrides the config log level
- `--pprof` - enables a profiling server on `localhost:6060` (off by default)

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/p2p-org/dkc/issues).

## License

[License](./LICENSE)
