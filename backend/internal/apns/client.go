package apns

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
)

type Client struct {
	clients map[string]*apns2.Client
	mu      sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		clients: make(map[string]*apns2.Client),
	}
}

// GetClient returns or creates an APNs client for the given credentials
func (c *Client) GetClient(keyPath, keyID, teamID, environment string) (*apns2.Client, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", teamID, keyID, environment, keyPath)

	c.mu.RLock()
	if client, exists := c.clients[cacheKey]; exists {
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	// Create new client
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := c.clients[cacheKey]; exists {
		return client, nil
	}

	authKey, err := loadAuthKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth key: %w", err)
	}

	jwtToken := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}

	client := apns2.NewTokenClient(jwtToken)

	if environment == "production" {
		client = client.Production()
	} else {
		client = client.Development()
	}

	c.clients[cacheKey] = client
	return client, nil
}

// loadAuthKey loads the ECDSA private key from a .p8 file
func loadAuthKey(path string) (*ecdsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA")
	}

	return ecdsaKey, nil
}

// Close closes all APNs clients
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, client := range c.clients {
		if client.HTTPClient != nil && client.HTTPClient.CloseIdleConnections != nil {
			client.HTTPClient.CloseIdleConnections()
		}
	}

	c.clients = make(map[string]*apns2.Client)
}
