package keyring

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/encryption/cloudkms"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	mu          sync.RWMutex
	defaultRing Ring
)

func Default() Ring {
	mu.RLock()
	defer mu.RUnlock()
	return defaultRing
}

// SetDefault overrides the default keyring.
// Note: This function is defined for testing purpose.
// Use Init to correctly setup a keyring.
func SetDefault(r Ring) {
	mu.Lock()
	defer mu.Unlock()
	defaultRing = r
}

func Init(ctx context.Context) error {
	config := conf.Get().EncryptionKeys
	ring, err := NewRing(ctx, config)
	if err != nil {
		return err
	}
	if ring != nil {
		defaultRing = *ring
	}

	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		if _, err := NewRing(ctx, cfg.EncryptionKeys); err != nil {
			return conf.Problems{conf.NewSiteProblem(fmt.Sprintf("Invalid encryption.keys config: %s", err))}
		}
		return nil
	})

	conf.Watch(func() {
		newConfig := conf.Get().EncryptionKeys
		if newConfig == config {
			return
		}
		newRing, err := NewRing(ctx, newConfig)
		if err != nil {
			panic("creating encryption keyring: " + err.Error())
		}
		mu.Lock()
		defaultRing = *newRing
		mu.Unlock()
	})
	return nil
}

// NewRing creates a keyring.Ring containing all the keys configured in site config
func NewRing(ctx context.Context, keyConfig *schema.EncryptionKeys) (*Ring, error) {
	if keyConfig == nil {
		return nil, nil
	}

	var r Ring
	var err error

	if keyConfig.ExternalServiceKey != nil {
		r.ExternalServiceKey, err = NewKey(ctx, keyConfig.ExternalServiceKey)
		if err != nil {
			return nil, err
		}
	}

	if keyConfig.UserExternalAccountKey != nil {
		r.UserExternalAccountKey, err = NewKey(ctx, keyConfig.UserExternalAccountKey)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type Ring struct {
	ExternalServiceKey     encryption.Key
	UserExternalAccountKey encryption.Key
}

func NewKey(ctx context.Context, k *schema.EncryptionKey) (encryption.Key, error) {
	if k == nil {
		return nil, fmt.Errorf("cannot configure nil key")
	}
	switch {
	case k.Cloudkms != nil:
		return cloudkms.NewKey(ctx, k.Cloudkms.Keyname)
	case k.Noop != nil:
		return &encryption.NoopKey{}, nil
	case k.Base64 != nil:
		return &encryption.Base64Key{}, nil
	default:
		return nil, fmt.Errorf("couldn't configure key: %v", *k)
	}
}

// MaybeEncrypt encrypts data with the given key returns the id of the key. If the key is nil, it returns the data unchanged.
func MaybeEncrypt(ctx context.Context, key encryption.Key, data string) (maybeEncryptedData, keyID string, err error) {
	var keyIdent string

	if key != nil {
		encrypted, err := key.Encrypt(ctx, []byte(data))
		if err != nil {
			return "", "", err
		}
		data = string(encrypted)
		keyIdent, err = key.ID(ctx)
		if err != nil {
			return "", "", err
		}
	}

	return data, keyIdent, nil
}

// MaybeDecrypt decrypts data with the given key if keyIdent is not empty.
func MaybeDecrypt(ctx context.Context, key encryption.Key, data, keyIdent string) (string, error) {
	if keyIdent == "" {
		// data is not encrypted, return plaintext
		return data, nil
	}
	if key == nil {
		return data, fmt.Errorf("couldn't decrypt encrypted data, key is nil")
	}
	decrypted, err := key.Decrypt(ctx, []byte(data))
	if err != nil {
		return data, err
	}

	return decrypted.Secret(), nil
}
