package apple

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const maximumKeyPayload = 1 << 20

type jsonWebKey struct {
	KeyID    string `json:"kid"`
	KeyType  string `json:"kty"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

type keySet struct {
	Keys []jsonWebKey `json:"keys"`
}

type keyCache struct {
	endpoint    string
	client      *http.Client
	ttl         time.Duration
	minimumWait time.Duration

	mutex     sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func newKeyCache(endpoint string, client *http.Client, ttl, minimumWait time.Duration) *keyCache {
	return &keyCache{endpoint: endpoint, client: client, ttl: ttl, minimumWait: minimumWait, keys: map[string]*rsa.PublicKey{}}
}

func (c *keyCache) key(ctx context.Context, keyID string, now time.Time) (*rsa.PublicKey, error) {
	c.mutex.Lock()
	cached, found := c.keys[keyID]
	fresh := now.Sub(c.fetchedAt) < c.ttl
	canRefresh := now.Sub(c.fetchedAt) >= c.minimumWait
	c.mutex.Unlock()

	if found && fresh {
		return cached, nil
	}
	if !canRefresh {
		if found {
			return cached, nil
		}
		return nil, ErrInvalidToken
	}
	refreshed, err := c.fetch(ctx)
	if err != nil {
		if found {
			return cached, nil
		}
		return nil, err
	}
	c.mutex.Lock()
	c.keys = refreshed
	c.fetchedAt = now
	key, ok := refreshed[keyID]
	c.mutex.Unlock()
	if !ok {
		return nil, ErrInvalidToken
	}
	return key, nil
}

func (c *keyCache) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return nil, errors.New("apple: key request could not be built")
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.New("apple: key retrieval failed")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("apple: key endpoint rejected the request")
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, maximumKeyPayload))
	if err != nil {
		return nil, errors.New("apple: key payload could not be read")
	}
	var document keySet
	if err := json.Unmarshal(payload, &document); err != nil {
		return nil, errors.New("apple: key payload was malformed")
	}
	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, candidate := range document.Keys {
		key, err := candidate.publicKey()
		if err != nil {
			continue
		}
		keys[candidate.KeyID] = key
	}
	if len(keys) == 0 {
		return nil, errors.New("apple: key payload contained no usable keys")
	}
	return keys, nil
}

func (k jsonWebKey) publicKey() (*rsa.PublicKey, error) {
	if k.KeyType != "RSA" || k.KeyID == "" {
		return nil, ErrInvalidToken
	}
	modulus, err := base64.RawURLEncoding.DecodeString(k.Modulus)
	if err != nil || len(modulus) == 0 {
		return nil, ErrInvalidToken
	}
	exponent, err := base64.RawURLEncoding.DecodeString(k.Exponent)
	if err != nil || len(exponent) == 0 || len(exponent) > 8 {
		return nil, ErrInvalidToken
	}
	value := new(big.Int).SetBytes(exponent)
	if !value.IsInt64() || value.Int64() < 3 {
		return nil, ErrInvalidToken
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(value.Int64())}, nil
}
