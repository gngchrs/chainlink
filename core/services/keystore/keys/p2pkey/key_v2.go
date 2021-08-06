package p2pkey

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"

	cryptop2p "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

type Raw []byte

func (rawKey Raw) Key() KeyV2 {
	privKey, err := cryptop2p.UnmarshalPrivateKey(rawKey)
	if err != nil {
		panic(err) // TODO - RYAN - uhhhmmmm?
	}
	key, err := fromPrivkey(privKey)
	if err != nil {
		panic(err)
	}
	return key
}

type KeyV2 struct {
	cryptop2p.PrivKey // TODO - RYAN embed?
	peerID            PeerID
}

func NewV2() (KeyV2, error) {
	privKey, _, err := cryptop2p.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return KeyV2{}, nil
	}
	return fromPrivkey(privKey)
}

func MustNewV2XXXTestingOnly(k *big.Int) KeyV2 {
	var privKeyBytes [64]byte
	copy(privKeyBytes[:], k.Bytes())
	p2pPrivKey, err := cryptop2p.UnmarshalEd25519PrivateKey(privKeyBytes[:])
	if err != nil {
		panic(err)
	}
	key, err := fromPrivkey(p2pPrivKey)
	if err != nil {
		panic(err)
	}
	return key
}

func (key KeyV2) ID() string {
	return peer.ID(key.peerID).String()
}

func (key KeyV2) Raw() Raw {
	marshalledPrivK, err := cryptop2p.MarshalPrivateKey(key.PrivKey)
	if err != nil {
		panic(err)
	}
	return marshalledPrivK
}

func (k KeyV2) PeerID() PeerID {
	return k.peerID
}

func (k KeyV2) PublicKeyHex() string {
	pubKeyBytes, err := k.GetPublic().Raw()
	if err != nil {
		panic(err) // TODO - RYAN
	}
	return hex.EncodeToString(pubKeyBytes)
}

func fromPrivkey(privKey cryptop2p.PrivKey) (KeyV2, error) {
	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return KeyV2{}, err
	}
	return KeyV2{
		PrivKey: privKey,
		peerID:  PeerID(peerID),
	}, nil
}
