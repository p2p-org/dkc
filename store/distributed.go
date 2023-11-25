package store

import (
	"context"
	"fmt"
	"regexp"

	"github.com/p2p-org/dkc/crypto/bls"
	"github.com/p2p-org/dkc/utils"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"
	"golang.org/x/exp/maps"
)

type Peers map[uint64]Peer
type Peer struct {
	Name        string
	Passphrases struct {
		Path  string
		Index int
	}
}

type Threshold uint32
type DistributedStore struct {
	Type           string
	Path           string
	Peers          map[uint64]string
	PeersPasswords map[uint64][][]byte
	PeersPaths     map[uint64]string
	Threshold      Threshold
	Ctx            context.Context
}

func (s *DistributedStore) Create() error {
	for id := range s.Peers {
		_, err := createStore(s.PeersPaths[id])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DistributedStore) GetWalletsAccountsMap() ([]AccountsData, []string, error) {
	// We assume that all distributed wallets in the store have the same accounts
	peers := maps.Values(s.Peers)
	res, err := regexp.Compile(`:.*`)
	if err != nil {
		return nil, nil, err
	}
	a, w, err := getWalletsAccountsMap(s.Ctx, s.Path+"/"+res.ReplaceAllString(peers[0], ""))
	if err != nil {
		return nil, nil, err
	}

	return a, w, nil
}

func (s *DistributedStore) CreateWallet(name string) error {
	for id := range s.Peers {
		store, err := getStore(s.PeersPaths[id])
		if err != nil {
			return err
		}
		_, err = e2wallet.CreateWallet(name, e2wallet.WithType(s.Type), e2wallet.WithStore(store))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DistributedStore) GetPK(w string, a string) ([]byte, error) {
	accounts := map[uint64][]byte{}

	for id := range s.Peers {
		wallet, err := getWallet(s.PeersPaths[id], w)
		if err != nil {
			return nil, err
		}
		err = wallet.(types.WalletLocker).Unlock(s.Ctx, nil)
		if err != nil {
			return nil, err
		}

		defer func() {
			err = wallet.(types.WalletLocker).Lock(s.Ctx)
		}()

		account, err := wallet.(types.WalletAccountByNameProvider).AccountByName(s.Ctx, a)
		if err != nil {
			return nil, err
		}

		key, err := getAccountPK(account, s.Ctx, s.PeersPasswords[id])
		if err != nil {
			return nil, err
		}

		accounts[id] = key

	}
	key, err := bls.Combine(s.Ctx, accounts)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *DistributedStore) SavePKToWallet(w string, a []byte, n string) error {
	// Spliting PK to shards and get Public and Private Keys for each shard
	masterSKs, masterPKs, err := bls.Split(s.Ctx, a, uint32(s.Threshold))
	if err != nil {
		return err
	}

	peersIDs := maps.Keys(s.Peers)
	participants, err := bls.SetupParticipants(masterSKs, masterPKs, peersIDs, len(s.Peers))
	if err != nil {
		return err
	}

	for id := range s.Peers {
		wallet, err := getWallet(s.PeersPaths[id], w)
		if err != nil {
			return err
		}
		err = wallet.(types.WalletLocker).Unlock(s.Ctx, nil)
		if err != nil {
			return err
		}

		defer func() {
			err = wallet.(types.WalletLocker).Lock(s.Ctx)
		}()

		_, err = wallet.(types.WalletDistributedAccountImporter).ImportDistributedAccount(
			s.Ctx,
			n,
			participants[id],
			uint32(s.Threshold),
			masterPKs,
			s.Peers,
			//Always Use The First Password In Array
			s.PeersPasswords[id][0],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func newDistributedStore(t string) (DistributedStore, error) {
	s := DistributedStore{}
	//Parse Wallet Type

	wt := viper.GetString(fmt.Sprintf("%s.wallet.type", t))

	utils.Log.Debug().Msgf("setting store type to %s", wt)
	s.Type = wt

	//Parse Store Path
	storePath := viper.GetString(fmt.Sprintf("%s.store.path", t))
	utils.Log.Debug().Msgf("setting store path to %s", storePath)
	if storePath == "" {
		return s, errors.New("distributed store path is empty")
	}
	s.Path = storePath

	//Parse Peers
	var peers Peers
	utils.Log.Debug().Msgf("getting peers")
	err := viper.UnmarshalKey(fmt.Sprintf("%s.wallet.peers", t), &peers)
	if err != nil {
		return s, err
	}

	//Peers list must be >= 2
	utils.Log.Debug().Msgf("checking peers length: %d", len(peers))
	if len(peers) < 2 {
		return s, errors.New("number of peers for distributed store is less than 2")
	}

	// Parse Peers Passwords Paths and Names
	// Regexp To Get Peers Path
	res, err := regexp.Compile(`:.*`)
	if err != nil {
		return s, err
	}
	s.Peers = map[uint64]string{}
	s.PeersPasswords = map[uint64][][]byte{}
	s.PeersPaths = map[uint64]string{}
	for id, peer := range peers {
		//Parse Passphrases
		utils.Log.Debug().Msgf("getting passhphrases for peers %s", peer.Name)
		passphrases, err := getAccountsPasswords(peer.Passphrases.Path)
		if err != nil {
			return s, err
		}
		utils.Log.Debug().Msgf("checking passhphrases len: %d for peers %s", len(passphrases), peer.Name)
		if len(passphrases) == 0 {
			return s, errors.New("passhparases file for distributed peer is empty")
		}
		// Cheking If Passphrases Index Is Set
		if viper.IsSet(fmt.Sprintf("%s.peers.%d.passphrases.index", t, id)) {
			passphrases = [][]byte{passphrases[viper.GetInt(fmt.Sprintf("%s.peers.%d.passphrases.index", t, id))]}
		}
		s.Peers[id] = peer.Name
		s.PeersPasswords[id] = passphrases
		s.PeersPaths[id] = s.Path + "/" + res.ReplaceAllString(peer.Name, "")
	}

	//Parse Threshold
	var threshold Threshold
	utils.Log.Debug().Msgf("getting threshold value")
	err = viper.UnmarshalKey(fmt.Sprintf("%s.wallet.threshold", t), &threshold)
	if err != nil {
		return s, err
	}

	//Check number of peers and threshold
	utils.Log.Debug().Msgf("checking threshold value")
	if uint32(threshold) <= uint32(len(peers)/2) {
		return s, errors.New("thershold value for distributed store is less than peers/2")
	}
	if uint32(threshold) > uint32(len(peers)) {
		return s, errors.New("threshold value for distributed store is more than peer")
	}

	s.Threshold = threshold

	return s, nil
}

func (s *DistributedStore) GetPath() string {
	return s.Path
}

func (s *DistributedStore) GetType() string {
	return s.Type
}
