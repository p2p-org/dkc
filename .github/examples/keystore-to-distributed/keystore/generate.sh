#!/bin/bash

# Generate wallets
ethdo  wallet create --base-dir . --wallet="Wallet1" --type="keystore" --wallet-passphrase="hdpass1" --allow-weak-passphrases
ethdo  wallet create --base-dir . --wallet="Wallet2" --type="keystore" --wallet-passphrase="hdpass2" --allow-weak-passphrases
ethdo  wallet create --base-dir . --wallet="Wallet3" --type="keystore" --wallet-passphrase="hdpass3" --allow-weak-passphrases


# Generate accounts
ethdo account create --base-dir . --account="Wallet1/Account1" --passphrase "hdaccountpass1" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet1/Account2" --passphrase "hdaccountpass1" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet1/Account3" --passphrase "hdaccountpass1" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet2/Account1" --passphrase "hdaccountpass2" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet2/Account2" --passphrase "hdaccountpass2" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet2/Account3" --passphrase "hdaccountpass2" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet3/Account1" --passphrase "hdaccountpass3" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet3/Account2" --passphrase "hdaccountpass3" --allow-weak-passphrases
ethdo account create --base-dir . --account="Wallet3/Account3" --passphrase "hdaccountpass3" --allow-weak-passphrases
