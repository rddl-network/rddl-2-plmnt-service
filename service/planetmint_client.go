package service

import (
	"context"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/types"

	"github.com/planetmint/planetmint-go/lib"
	daotypes "github.com/planetmint/planetmint-go/x/dao/types"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type IPlanetmintClient interface {
	MintPLMNT(beneficiary string, amount uint64, liquidTxHash string) (err error)
	CheckMintRequest(txhash string) (mintRequest *daotypes.QueryGetMintRequestsByHashResponse, err error)
}

type PlanetmintClient struct{}

func NewPlanetmintClient() *PlanetmintClient {
	return &PlanetmintClient{}
}

func (pmc *PlanetmintClient) MintPLMNT(beneficiary string, amount uint64, liquidTxHash string) (err error) {
	cfg := config.GetConfig()
	mintRequest := daotypes.MintRequest{
		Beneficiary:  beneficiary,
		Amount:       amount,
		LiquidTxHash: liquidTxHash,
	}

	addr := types.MustAccAddressFromBech32(cfg.PlanetmintAddress)
	msg := daotypes.NewMsgMintToken(cfg.PlanetmintAddress, &mintRequest)

	_, err = lib.BroadcastTxWithFileLock(addr, msg)
	if err != nil {
		return
	}

	return
}

func (pmc *PlanetmintClient) CheckMintRequest(txhash string) (mintRequest *daotypes.QueryGetMintRequestsByHashResponse, err error) {
	cfg := config.GetConfig()
	grcpConn, err := grpc.Dial(
		cfg.PlanetmintRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return
	}

	daoClient := daotypes.NewQueryClient(grcpConn)
	mintRequest, err = daoClient.GetMintRequestsByHash(
		context.Background(),
		&daotypes.QueryGetMintRequestsByHashRequest{Hash: txhash},
	)

	if err != nil {
		if strings.Contains(err.Error(), codes.NotFound.String()) {
			err = nil
		}
		return
	}

	return
}

// verifyAddress verifies the integrity and prefix of a given address.
func VerifyAddress(address string) (valid bool, err error) {
	// Attempt to decode the address
	_, err = types.AccAddressFromBech32(address)
	if err != nil {
		return
	}
	if !strings.Contains(address, "plmnt") {
		valid = false
		return
	}
	valid = true
	return
}
