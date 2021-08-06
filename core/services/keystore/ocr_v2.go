package keystore

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/ocrkey"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/p2pkey"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

type OCR interface {
	// P2P Keys
	GenerateP2PKey() (p2pkey.KeyV2, error)
	AddP2PKey(p2pkey.KeyV2) error
	GetP2PKeys() (keys []p2pkey.KeyV2, err error)
	GetP2PKey(id string) (*p2pkey.KeyV2, error)
	DeleteP2PKey(key *p2pkey.KeyV2) error
	ImportP2PKey(keyJSON []byte, password string) (*p2pkey.KeyV2, error)
	ExportP2PKey(id string, password string) ([]byte, error)
	// OCR Keys
	GenerateOCRKey() (ocrkey.KeyV2, error)
	AddOCRKey(ocrkey.KeyV2) error
	GetOCRKeys() ([]ocrkey.KeyV2, error)
	GetOCRKey(id string) (ocrkey.KeyV2, error)
	DeleteOCRKey(id string) error
	ImportOCRKey(keyJSON []byte, password string) (*ocrkey.KeyV2, error)
	ExportOCRKey(id string, password string) ([]byte, error)
}

type ocr struct {
	*keyManager
}

var _ OCR = ocr{}

func newOCRKeyStore(km *keyManager) ocr {
	return ocr{
		km,
	}
}

func (ks ocr) GenerateP2PKey() (p2pkey.KeyV2, error) {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return p2pkey.KeyV2{}, LockedErr
	}
	key, err := p2pkey.NewV2()
	if err != nil {
		return p2pkey.KeyV2{}, errors.Wrapf(err, "while generating new p2p key")
	}
	err = ks.safeAddKey(key)
	if err != nil {
		return p2pkey.KeyV2{}, errors.Wrapf(err, "while adding new p2p key")
	}
	return key, nil
}

func (ks ocr) AddP2PKey(key p2pkey.KeyV2) error {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return LockedErr
	}
	return ks.safeAddKey(key)
}

func (ks ocr) GetP2PKeys() (keys []p2pkey.KeyV2, err error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return keys, LockedErr
	}
	for _, key := range ks.keyRing.P2P {
		keys = append(keys, key)
	}
	return keys, nil
}

func (ks ocr) GetP2PKey(id string) (*p2pkey.KeyV2, error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return nil, LockedErr
	}
	return ks.getP2PKey(id)
}

func (ks ocr) DeleteP2PKey(key *p2pkey.KeyV2) error {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return LockedErr
	}
	return ks.safeRemoveKey(*key)
}

func (ks ocr) ImportP2PKey(keyJSON []byte, password string) (*p2pkey.KeyV2, error) {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return nil, LockedErr
	}
	key, err := p2pkey.FromEncryptedJSON(keyJSON, password)
	if err != nil {
		return nil, errors.Wrap(err, "P2PKeyStore#ImportKey failed to decrypt key")
	}
	return &key, ks.safeAddKey(key)
}

func (ks ocr) ExportP2PKey(id string, password string) ([]byte, error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return nil, LockedErr
	}
	key, err := ks.getP2PKey(id)
	if err != nil {
		return nil, err
	}
	return key.ToEncryptedJSON(password, ks.scryptParams)
}

func (ks ocr) getP2PKey(id string) (*p2pkey.KeyV2, error) {
	key, found := ks.keyRing.P2P[id]
	if !found {
		return nil, errors.New(fmt.Sprintf("P2P key not found with ID %s", id))
	}
	return &key, nil
}

// TODO - change this signature to accept key ID type
func (ks ocr) DecryptedOCRKey(hash models.Sha256Hash) (ocrkey.KeyV2, bool) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	keyID := hex.EncodeToString(hash[:])
	k, exists := ks.keyRing.OCR[keyID]
	if !exists {
		return ocrkey.KeyV2{}, false
	}
	return k, true
}

func (ks ocr) GenerateOCRKey() (ocrkey.KeyV2, error) {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return ocrkey.KeyV2{}, LockedErr
	}
	key, err := ocrkey.NewV2()
	if err != nil {
		return ocrkey.KeyV2{}, errors.Wrapf(err, "while generating new ocr key")
	}
	err = ks.safeAddKey(key)
	if err != nil {
		return ocrkey.KeyV2{}, errors.Wrapf(err, "while adding new ocr key")
	}
	return key, nil
}

func (ks ocr) AddOCRKey(key ocrkey.KeyV2) error {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return LockedErr
	}
	return ks.safeAddKey(key)
}

func (ks ocr) GetOCRKeys() (keys []ocrkey.KeyV2, err error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return keys, LockedErr
	}
	for _, key := range ks.keyRing.OCR {
		keys = append(keys, key)
	}
	return keys, nil
}

func (ks ocr) GetOCRKey(id string) (ocrkey.KeyV2, error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return ocrkey.KeyV2{}, LockedErr
	}
	return ks.getOCRKey(id)
}

func (ks ocr) DeleteOCRKey(id string) error {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return LockedErr
	}
	key, err := ks.getOCRKey(id)
	if err != nil {
		return err
	}
	return ks.safeRemoveKey(key)
}

func (ks ocr) ImportOCRKey(keyJSON []byte, password string) (*ocrkey.KeyV2, error) {
	ks.lock.Lock()
	defer ks.lock.Unlock()
	if ks.isLocked() {
		return nil, LockedErr
	}
	key, err := ocrkey.FromEncryptedJSON(keyJSON, password)
	if err != nil {
		return nil, errors.Wrap(err, "OCRKeyStore#ImportKey failed to decrypt key")
	}
	return &key, ks.safeAddKey(key)
}

func (ks ocr) ExportOCRKey(id string, password string) ([]byte, error) {
	ks.lock.RLock()
	defer ks.lock.RUnlock()
	if ks.isLocked() {
		return nil, LockedErr
	}
	key, err := ks.getOCRKey(id)
	if err != nil {
		return nil, err
	}
	return key.ToEncryptedJSON(password, ks.scryptParams)
}

// caller must hold lock
func (ks ocr) getOCRKey(id string) (ocrkey.KeyV2, error) {
	key, found := ks.keyRing.OCR[id]
	if !found {
		return ocrkey.KeyV2{}, errors.New(fmt.Sprintf("OCR key not found with ID %s", id))
	}
	return key, nil
}
