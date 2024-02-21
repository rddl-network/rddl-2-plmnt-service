package testutil

import (
	machinetypes "github.com/planetmint/planetmint-go/x/machine/types"
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
	IsLegitMachine              = machinetypes.QueryGetMachineByAddressResponse{Machine: &machine}
	machine                     = machinetypes.Machine{Name: PlanetmintAddress, Ticker: "", Domain: "", Reissue: false, Amount: 1, Precision: 8, IssuerPlanetmint: "", IssuerLiquid: "", MachineId: "", MachineIdSignature: "", Metadata: &metadata, Type: 1, Address: PlanetmintAddress}
	PlanetmintAddress           = "plmnt1683t0us0r85840nsepx6jrk2kjxw7zrcnkf0rp"
	metadata                    = machinetypes.Metadata{Gps: gps, Device: device, AssetDefinition: "{\"Version\":\"0.2\"}", AdditionalDataCID: ""}
	gps                         = "{\"Country\":\"AT\",\"Region\":\"9\",\"City\":\"vienna\",\"CityLatLong\":\"48.208174,16.373819\"}"
	device                      = "{\"Category\":\"MyMachine\", \"Manufacturer\":\"DIY\"}"
	ConfidentialAddr            = "tlq1qqt2tw28n29t6jcdspnz2nc4cqack596wryvuvjm3w3fey3a572flxjvy3xu6kd4nmx8hs8fzq9ns3vr9e7q0s22cu2pp7m2l4"
	UnconfidentialAddr          = "tex1qfxzgnwdtx6eanrmcr53qzecgkpjulq8crkueph"
	AddressInfo                 = types.GetAddressInfoResult{Confidential: UnconfidentialAddr}
)
