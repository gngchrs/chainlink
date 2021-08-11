package ocr2key

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ ocrtypes.OnchainKeyring = &EthereumKeyring{}

type EthereumKeyring struct {
	privateKey ecdsa.PrivateKey
}

func NewEthereumKeyring(material io.Reader) (*EthereumKeyring, error) {
	ecdsaKey, err := ecdsa.GenerateKey(curve, material)
	if err != nil {
		return nil, err
	}
	return &EthereumKeyring{privateKey: *ecdsaKey}, nil
}

// XXX: PublicKey returns the address of the public key not the public key itself
func (ok *EthereumKeyring) PublicKey() ocrtypes.OnchainPublicKey {
	publicKey := (*ecdsa.PrivateKey)(&ok.privateKey).Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)
	return address[:]
}

func (ok *EthereumKeyring) Sign(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, &reportCtx)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, &report)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(onChainHash(buffer.Bytes()), (*ecdsa.PrivateKey)(&ok.privateKey))
}

func (ok *EthereumKeyring) Verify(publicKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report ocrtypes.Report, signature []byte) bool {
	digest, err := ok.Sign(reportCtx, report)
	if err != nil {
		// FIXME: this is really an error not false but the interface limits us
		return false
	}
	return crypto.VerifySignature(publicKey, digest, signature)
}

func (ok *EthereumKeyring) MaxSignatureLength() int {
	return 65
}

func (ok *EthereumKeyring) SigningAddress() common.Address {
	return crypto.PubkeyToAddress(*(*ecdsa.PrivateKey)(&ok.privateKey).Public().(*ecdsa.PublicKey))
}

func (ok *EthereumKeyring) marshal() ([]byte, error) {
	return crypto.FromECDSA(&ok.privateKey), nil
}

func (ok *EthereumKeyring) unmarshal(in []byte) error {
	privateKey, err := crypto.ToECDSA(in)
	ok.privateKey = *privateKey
	return err
}

func onChainHash(msg []byte) []byte {
	return crypto.Keccak256(msg)
}
