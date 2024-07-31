package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	PrivateKey *ecdsa.PrivateKey
	USD        float64
}

func NewUser(pk *ecdsa.PrivateKey, usd float64) *User {
	return &User{
		ID:         uuid.New(),
		USD:        usd,
		PrivateKey: pk,
	}
}

func GenerateUsers() []*User {
	privKeys := make([]*ecdsa.PrivateKey, 0)
	strings := []string{
		"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6",
		"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		"5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
	}

	for _, s := range strings {
		pk, err := crypto.HexToECDSA(s)
		if err != nil {
			fmt.Printf("Error converting hex to ECDSA: %v\n", err)
			continue
		}
		privKeys = append(privKeys, pk)
	}
	users := make([]*User, 0)

	for i := 0; i < 3; i++ {
		user := NewUser(privKeys[i], 0)
		users = append(users, user)

	}


	return users
}

func GetBalance(user *User, client *ethclient.Client) float64 {
	addr := internals.GetAddress(user.PrivateKey)
	bal, _ := client.BalanceAt(context.Background(), addr, nil)

	ans, _ := internals.WeiToEther(bal).Float64()
	return ans
}
