package auth

import (
	"context"
	"crypto/ecdsa"

	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(pk *ecdsa.PrivateKey) *User {
	return &User{
		ID: uuid.New(),
	}
}

func GenerateUsers() []*User {
	privKeys := make([]*ecdsa.PrivateKey, 0)
	strings := []string{
		"0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
		"0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		"0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
	}

	for _, s := range strings {
		pk, _ := crypto.HexToECDSA(s)
		privKeys = append(privKeys, pk)
	}
	users := make([]*User, 0)

	for i := 0; i < 3; i++ {
		users = append(users, NewUser(privKeys[i]))
	}

	return users
}

func GetBalance(user *User, client *ethclient.Client) float64 {
	addr := internals.GetAddress(user.PrivateKey)
	bal, _ := client.BalanceAt(context.Background(), addr, nil)

	ans, _ := internals.WeiToEther(bal).Float64()
	return ans
}
