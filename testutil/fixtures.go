package testutil

import "github.com/rddl-network/elements-rpc/types"

var (
	GetTransactionResult = types.GetTransactionResult{Hex: zeros, TxID: zeros, Amount: map[string]float64{
		"7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9": 10000,
	}}
	zeros = "0000000000000000000000000000000000000000000000000000000000000000"
)
