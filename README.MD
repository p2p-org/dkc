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

`dkc` uses [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary). Some research code of using [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary) in `dkc` is available [here](./research/research.go) 

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
  path: nd_wallets                  #Path to non-determenistic wallet (Default: None)
  passphrases: ./passphrases.txt    #Path to file containing passphrases for unlocking/locking accounts (Default: None)
log-level: debug                    #Log-level (Default: INFO)
```

### File structure

##### Combine

The following is an [example](.github/examples/wallets) file structure if one were to combine threshold keys. `old1`, `old2`, `old3` are each the `base-dir` wallet directory from the dirk instance. `old1`, `old2`, `old3` are each their own wallet directory. 

`ethdo wallet list --base-dir ./wallet` should return nothing.

The subdirectories of `wallet` folder are the actual `ethdo` wallets: 

```
$ ethdo wallet list --base-dir wallets/old1 
9faf1408-839c-4b21-9d15-80d192c2bced
$ ethdo wallet list --base-dir wallets/old2
ff3918f1-16ef-47bf-83d2-16b5a64c9aa1
$ ethdo wallet list --base-dir wallets/old3 
dd428711-4681-41d1-ad47-746cc6cfea8f
```

It should look like this:

```
wallet
├── old1
│   └── 9a76c4f3-aa46-4864-b7fd-55681f8afa3b
│       ├── 69e976d2-3c78-44e3-b9a6-45149078723d
│       ├── 85587167-c356-4968-aa36-37464d05640c
│       ├── 9a76c4f3-aa46-4864-b7fd-55681f8afa3b
│       ├── a99d9d83-4afb-4870-8cce-a57c69c6bda2
│       ├── fd83339a-d36e-4c7a-a30c-0999be075dcb
│       └── index
├── old2
│   └── 32b277e3-b71a-45ed-b47d-5b6c155b8cf9
│       ├── 026b1c48-9a1a-49f8-b289-792d79378090
│       ├── 32b277e3-b71a-45ed-b47d-5b6c155b8cf9
│       ├── 6880aa1f-e490-4508-868f-f2af3b7f34f7
│       ├── ab1cc0d8-5447-41bd-b3b5-0c100ae9747a
│       ├── f282e1ff-0364-48e6-a674-26ea18184bf3
│       └── index
└── old3
    └── 68f2bf6c-4266-49e2-915a-be9946c0882b
        ├── 5613693b-1b36-4bac-9b85-d96deeeb8d7c
        ├── 5a8c805d-9056-4ded-b9a2-611cdaa99427
        ├── 5ee08d46-70d0-430a-b4ef-f461567c344c
        ├── 68f2bf6c-4266-49e2-915a-be9946c0882b
        ├── e177e6d7-195a-44a0-b350-7bcf73642253
        └── index
```

*Importantly, the names of each wallet folder must correspond with the values in `distributed-wallets.peers` defined in the config.*

The keys within each wallet must also have the same name

```
cat wallets/old1/56f23183-85e7-4f65-863e-738975f137ad/index | jq
[
  {
    "uuid": "37be3c86-5d24-47f3-99b8-83b6ac62e6b7",
    "name": "3"
  },
  {
    "uuid": "54799fa9-3239-4df2-857b-50c5d725cbfe",
    "name": "4"
  },
  {
    "uuid": "5676a9e5-e1de-4eef-9473-8c52c0810ce5",
    "name": "1"
  },
  {
    "uuid": "c153e6f1-2ae2-44fd-9a4b-e91965c8a2a6",
    "name": "2"
  }
]

cat wallets/old3/e5dc7bb0-9be2-45a5-b54b-768f4fd105df/index | jq
[
  {
    "uuid": "d38d8796-5022-42ea-8495-cd5d67743afc",
    "name": "3"
  },
  {
    "uuid": "0d69229d-cbcc-42eb-b29b-10ae91f28dff",
    "name": "4"
  },
  {
    "uuid": "b043fbf2-f392-4c09-b8ff-319ed8345dc8",
    "name": "1"
  },
  {
    "uuid": "c38ac2af-f5ee-4326-bd03-8572d2449623",
    "name": "2"
  }
]

cat wallets/old3/e5dc7bb0-9be2-45a5-b54b-768f4fd105df/index | jq
[
  {
    "uuid": "d38d8796-5022-42ea-8495-cd5d67743afc",
    "name": "3"
  },
  {
    "uuid": "0d69229d-cbcc-42eb-b29b-10ae91f28dff",
    "name": "4"
  },
  {
    "uuid": "b043fbf2-f392-4c09-b8ff-319ed8345dc8",
    "name": "1"
  },
  {
    "uuid": "c38ac2af-f5ee-4326-bd03-8572d2449623",
    "name": "2"
  }
]
```

All of the keys with corresponding names (ex: `name = 1`) should be the threshold keys corresponding to the same composite public key. 

##### Split

The following is an example file structure if one were to want to split non-deterministic keys into threshold keys.

```
nd_wallets
└── 20384d86-bc39-4ac4-a827-3cff41043cbd
    ├── 20384d86-bc39-4ac4-a827-3cff41043cbd
    ├── 2c2cd425-fa58-43b7-808b-d2ba32d692cc
    ├── 5bee4007-73ba-452f-9e53-6d04110dbb7c
    ├── 95b9898b-667f-49cd-9097-9da9bf760173
    ├── b9d79ea4-9e15-4f3c-9c63-b67e7965a060
    └── index
```

In this instance, `nd_wallets` is the wallet `base-dir`.
```
$ ethdo wallet list --base-dir nd_wallets
be7efee0-a219-4fb5-a0b7-a47662320755
```

To split up, non-deterministic keys, just put them into the wallet.

```
nd_wallets
└── 20384d86-bc39-4ac4-a827-3cff41043cbd
    ├── 20384d86-bc39-4ac4-a827-3cff41043cbd
    ├── 2c2cd425-fa58-43b7-808b-d2ba32d692cc
    ├── 5bee4007-73ba-452f-9e53-6d04110dbb7c
    ├── 95b9898b-667f-49cd-9097-9da9bf760173
    ├── b9d79ea4-9e15-4f3c-9c63-b67e7965a060
    └── index
```

##### Output Formats

The output file hierarchy is the same as the opposite side's input hierarchy. For example, the format needed to split keys, will be the output format of combined keys. 

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
