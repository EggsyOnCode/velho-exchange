package internals

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// Generate a new ECDSA private key
func GenerateNewPrivateKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error generating new private key: %v", err)
	}
	return privKey
}

func GetPrivKeyFromHexString(key string) (*ecdsa.PrivateKey, error) {
	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func GetAddress(privateKey *ecdsa.PrivateKey) common.Address {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	return addr
}

func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// etherToWei converts Ether to Wei.
func EtherToWei(ether *big.Float) *big.Int {
	// Conversion factor: 1 Ether = 10^18 Wei
	weiFactor := new(big.Float).SetFloat64(1e18)

	// Multiply ether value by the conversion factor to get Wei value
	weiValue := new(big.Float).Mul(ether, weiFactor)

	// Convert the resulting big.Float to big.Int
	weiInt, _ := weiValue.Int(nil)
	return weiInt
}

func NewEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetGasPrice() (float64, error) {
	client, _ := NewEthClient()
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0.0, err
	}

	val, _ := gasPrice.Float64()
	return val / 1e18, nil
}

// value will be sent in Wei
func TransferETH(from *ecdsa.PrivateKey, to common.Address, valueInETH float64) error {
	client, _ := NewEthClient()
	fromAddr := GetAddress(from)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		log.Fatal(err)
	}
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewFloat(valueInETH)
	wei := EtherToWei(value)

	var data []byte
	tx := types.NewTransaction(nonce, to, wei, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	// fmt.Printf("tx sent: %s \n value was : %v \n", signedTx.Hash().Hex(), wei)
	return nil
}

func GetBalance(addr common.Address) float64 {
	client, _ := NewEthClient()
	bal, _ := client.BalanceAt(context.Background(), addr, nil)

	ans, _ := WeiToEther(bal).Float64()
	return ans
}
