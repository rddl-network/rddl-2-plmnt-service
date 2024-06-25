package testutil

import (
	"github.com/rddl-network/elements-rpc/types"
)

var (
	GetTransactionResult = types.GetTransactionResult{Hex: zeros, TxID: zeros, Details: GetTransactionResultDetails, Amount: map[string]float64{
		"7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9": 10000,
	}, Confirmations: 10}
	DeriveAddressesResult       = types.DeriveAddressesResult{"tex1q2xn886usv9wxuvnfa6cll4ny6q2dk99mcpefk4"}
	GetTransactionResultDetails = []types.GetTransactionDetailsResult{{Address: "tex1q2xn886usv9wxuvnfa6cll4ny6q2dk99mcpefk4"}}
	zeros                       = "0000000000000000000000000000000000000000000000000000000000000000"
	GetNewAddress               = "tex1q2xn886usv9wxuvnfa6cll4ny6q2dk99mcpefk4"
	PlanetmintAddress           = "plmnt1683t0us0r85840nsepx6jrk2kjxw7zrcnkf0rp"
	ConfidentialAddr            = "tlq1qqt2tw28n29t6jcdspnz2nc4cqack596wryvuvjm3w3fey3a572flxjvy3xu6kd4nmx8hs8fzq9ns3vr9e7q0s22cu2pp7m2l4"
	UnconfidentialAddr          = "tex1qfxzgnwdtx6eanrmcr53qzecgkpjulq8crkueph"
	AddressInfo                 = types.GetAddressInfoResult{Confidential: UnconfidentialAddr}
	ReceivedTxByAddress1Tx      = types.ListReceivedByAddressResult{Address: ConfidentialAddr, Amount: 2.00000000, Confirmations: 2, TxIDs: []string{"44e7812ffa95a4031c1b97f534c2535fdad583627203bf63db8d5909902b6a87"}}
	ReceivedTxByAddress2Tx      = types.ListReceivedByAddressResult{Address: ConfidentialAddr, Amount: 2.00000000, Confirmations: 2, TxIDs: []string{"44e7812ffa95a4031c1b97f534c2535fdad583627203bf63db8d5909902b6a87", "87d8be31018183c7b6e013ef712d186a3e7aca08b37abe6bc86acda23692cb9b"}}
	ReceivedTxByAddressArray2Tx = []types.ListReceivedByAddressResult{ReceivedTxByAddress2Tx}
	ReceivedTxByAddressArray1Tx = []types.ListReceivedByAddressResult{ReceivedTxByAddress1Tx}
)
