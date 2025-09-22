package internals

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNewPrivateKey(t *testing.T) {
	privKey := GenerateNewPrivateKey()
	assert.NotNil(t, privKey, "Private key generation failed")
	assert.NotNil(t, privKey.PublicKey, "Public key generation failed")
}

func TestGetPrivKeyFromHexString(t *testing.T) {
	// Example private key (use a dummy one for testing)
	keyHex := "4c0883a69102937d6231471b5dbb6204fe512961708279b23456789aedaed123"
	privKey, err := GetPrivKeyFromHexString(keyHex)
	assert.Nil(t, err, "Error while converting private key from hex string")
	assert.NotNil(t, privKey, "Private key conversion returned nil")
}

func TestGetAddress(t *testing.T) {
	privKey := GenerateNewPrivateKey()
	address := GetAddress(privKey)
	assert.NotEqual(t, common.Address{}, address, "Address generation failed")
}

func TestWeiToEther(t *testing.T) {
	wei := big.NewInt(1e18)
	ether := WeiToEther(wei)
	expected := big.NewFloat(1.0)
	assert.Equal(t, 0, ether.Cmp(expected), "Wei to Ether conversion failed")
}

func TestEtherToWei(t *testing.T) {
	ether := big.NewFloat(1.0)
	wei := EtherToWei(ether)
	expected := big.NewInt(1e18)
	assert.Equal(t, 0, wei.Cmp(expected), "Ether to Wei conversion failed")
}

func TestNewEthClient(t *testing.T) {
	client, err := NewEthClient()
	assert.Nil(t, err, "Failed to create Ethereum client")
	assert.NotNil(t, client, "Ethereum client is nil")
}

func TestGetGasPrice(t *testing.T) {
	_, err := GetGasPrice()
	assert.Nil(t, err, "Error while fetching gas price")
}

// func TestTransferETH(t *testing.T) {
// 	// Mock private keys and addresses for testing
// 	fromPrivKey := GenerateNewPrivateKey()
// 	toAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

// 	err := TransferETH(fromPrivKey, toAddr, 0.01)
// 	assert.Nil(t, err, "Failed to transfer ETH")
// }

//	func TestGetBalance(t *testing.T) {
//		// Use a valid address for testing
//		addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
//		balance := GetBalance(addr)
//		assert.GreaterOrEqual(t, balance, 0.0, "Invalid balance fetched")
//	}
func FuzzGetPrivKeyFromHexString(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := GetPrivKeyFromHexString(string(data))
		if err != nil {
			// Check if the error is a hex decoding error
			_, hexErr := hex.DecodeString(string(data))
			if hexErr == nil && len(data) > 0 {
				// If hex.DecodeString doesn't return an error, it's an unexpected error
				t.Errorf("Unexpected error for input: %v, error: %v", string(data), err)
			}
		}
	})
}
