/**
 * Copyright 2019, Z Lab Corporation. All rights reserved.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package keymanager

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/crypto/ripemd160"
)

type KeyType int32

const (
	UNSPECIFIED_KEY_TYPE KeyType = 0
	EC_P256              KeyType = 1
	EC_P384              KeyType = 2
	RSA_2048             KeyType = 4
	RSA_4096             KeyType = 5
)

const (
	KeyAKey = "Key-A"
	KeyBKey = "Key-B"
)

type KeyManager interface {
	GetPrivateKey() (string, crypto.PrivateKey, error)
	GetPublicKeys() (map[string]crypto.PublicKey, error)
	GenerateKey() (*KeyPair, error)
	WriteEntry() error
}

type KeyPair struct {
	// Key pair identifier. It is used as 'kid' in JWK.
	ID string `json:"id"`
	// DER encoded private-key
	PrivateKeyDER []byte `json:"privateKey"`
	// DER encoded public-key
	PublicKeyDER []byte `json:"publicKey"`
	// Time when key pair was issued
	IssueAt time.Time `json:"issueAt"`
	// Time when key pair is expired
	ExpireAt time.Time `json:"expireAt"`
}

type DiskKeyManager struct {
	mu sync.RWMutex
	// Type of key
	Type KeyType
	// Path to a key-pair file
	Path string
	// Rotation period
	RotationPeriod time.Duration
	// Check key expiration interval
	Interval time.Duration
	// KeyPair data read from file which specified by Path
	KeyPairs map[string]KeyPair
	// Current entry to use sign data
	current string
	// A key-pair entry to use next time
	next string
}

func NewDiskKeyManager(t KeyType, p string, rp time.Duration, iv time.Duration) *DiskKeyManager {
	return &DiskKeyManager{
		mu:             sync.RWMutex{},
		Type:           t,
		Path:           p,
		RotationPeriod: rp,
		Interval:       iv,
	}
}

func (d *DiskKeyManager) Initialize() error {
	if _, err := os.Stat(d.Path); os.IsNotExist(err) {
		kp, err := d.GenerateKey()
		if err != nil {
			return err
		}
		d.KeyPairs = map[string]KeyPair{
			KeyAKey: *kp,
		}
		d.current = KeyAKey
		d.next = KeyBKey
		if err := d.WriteEntry(); err != nil {
			return err
		}

		go d.rotateKeys()

		return nil
	}

	b, err := ioutil.ReadFile(d.Path)
	if err != nil {
		return err
	}

	var data map[string]KeyPair
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	d.KeyPairs = data

	if d.KeyPairs[KeyAKey].IssueAt.Before(d.KeyPairs[KeyBKey].IssueAt) {
		d.current = KeyBKey
		d.next = KeyAKey
	} else {
		d.current = KeyAKey
		d.next = KeyBKey
	}

	glog.Info("KeyManager is Initialized")
	glog.V(4).Infof("Current Key: %v\n", d.current)

	go d.rotateKeys()

	return nil
}

func (d *DiskKeyManager) WriteEntry() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	b, err := json.MarshalIndent(d.KeyPairs, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(d.Path, b, 0600)
}

func (d *DiskKeyManager) GenerateKey() (*KeyPair, error) {
	var (
		privData []byte
		pubData  []byte
	)

	now := time.Now()

	switch d.Type {
	case EC_P256:
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		privData, err = x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		pubData, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return nil, err
		}
	case EC_P384:
		privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, err
		}
		privData, err = x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		pubData, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return nil, err
		}
	case RSA_2048:
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		privData = x509.MarshalPKCS1PrivateKey(privateKey)
		pubData, err = x509.MarshalPKIXPublicKey(privateKey.PublicKey)
		if err != nil {
			return nil, err
		}
	case RSA_4096:
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		privData = x509.MarshalPKCS1PrivateKey(privateKey)
		pubData, err = x509.MarshalPKIXPublicKey(privateKey.PublicKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("KeyType must be specified")
	}

	rip := ripemd160.New()
	_, err := io.WriteString(rip, string(pubData))
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		ID:            fmt.Sprintf("%x", rip.Sum(nil)),
		PrivateKeyDER: privData,
		PublicKeyDER:  pubData,
		IssueAt:       now,
		ExpireAt:      now.Add(d.RotationPeriod),
	}, nil
}

func (d *DiskKeyManager) GetPublicKeys() (map[string]crypto.PublicKey, error) {
	pubKeys := map[string]crypto.PublicKey{}

	for i := range d.KeyPairs {
		kp := d.KeyPairs[i]
		pub, err := x509.ParsePKIXPublicKey(kp.PublicKeyDER)
		if err != nil {
			return nil, err
		}
		pubKeys[kp.ID] = pub
	}
	return pubKeys, nil
}

func (d *DiskKeyManager) GetPrivateKey() (string, crypto.PrivateKey, error) {
	curKP := d.KeyPairs[d.current]
	curID := curKP.ID
	curPrivKey := curKP.PrivateKeyDER

	switch d.Type {
	case EC_P256, EC_P384:
		key, err := x509.ParseECPrivateKey(curPrivKey)
		return curID, key, err
	case RSA_2048, RSA_4096:
		key, err := x509.ParsePKCS1PrivateKey(curPrivKey)
		return curID, key, err
	default:
		return "", nil, errors.New("unknown key type")
	}
}

func (d *DiskKeyManager) shouldPrepareNewKey() bool {
	cur := d.KeyPairs[d.current]
	next := d.KeyPairs[d.next]
	lifetime := cur.ExpireAt.Sub(cur.IssueAt)
	glog.V(10).Infof("renew after: %v", cur.ExpireAt.Add(-lifetime/2))
	return time.Now().After(cur.ExpireAt.Add(-lifetime/2)) && !cur.IssueAt.Before(next.IssueAt)
}

func (d *DiskKeyManager) shouldRotateKey() bool {
	cur := d.KeyPairs[d.current]
	lifetime := cur.ExpireAt.Sub(cur.IssueAt)
	glog.V(10).Infof("rotate after: %v", cur.ExpireAt.Add(-lifetime/6))
	return time.Now().After(cur.ExpireAt.Add(-lifetime / 6))
}

func (d *DiskKeyManager) rotateKeys() {
	t := time.NewTicker(d.Interval)
	defer t.Stop()

	glog.Info("Start to rotate key-pairs loop.")

	for {
		select {
		case <-t.C:
			curKeyName := d.current

			if d.shouldPrepareNewKey() {
				glog.Infof("%v should be rotation", curKeyName)

				newKey, err := d.GenerateKey()
				if err != nil {
					glog.Errorf("Failed to generate keys for rotation: %v", err)
					continue
				}
				d.KeyPairs[d.next] = *newKey
				if err := d.WriteEntry(); err != nil {
					glog.Errorf("Failed to write keys: %v", err)
				}
				glog.Info("generate new key-pair before rotation")
			}

			if d.shouldRotateKey() {
				d.current = d.next
				d.next = curKeyName
				if err := d.WriteEntry(); err != nil {
					glog.Errorf("Failed to write keys: %v", err)
				}
				glog.Infof("%v is activated.", d.current)
			}
		}
	}
}
