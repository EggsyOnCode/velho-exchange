package internals

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func GetAddress(privateKey *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(privateKey.PublicKey)
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

// value will be sent in Wei
func TransferETH(from *ecdsa.PrivateKey, to common.Address, valueInETH float64, client *ethclient.Client) error {
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
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return nil
}
