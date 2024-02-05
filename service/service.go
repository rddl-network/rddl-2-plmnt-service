package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/planetmint/planetmint-go/lib"
	daotypes "github.com/planetmint/planetmint-go/x/dao/types"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type R2PService struct {
	router *gin.Engine
}

func NewR2PService(router *gin.Engine) *R2PService {
	service := &R2PService{router: router}
	service.registerRoutes()
	return service
}

func (r2p *R2PService) Run(config *viper.Viper) {
	bindAddress := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	_ = r2p.router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

func (r2p *R2PService) mintPLMNT(beneficiary string, amount uint64, liquidTxHash string) (err error) {
	cfg := config.GetConfig()
	mintRequest := daotypes.MintRequest{
		Beneficiary:  beneficiary,
		Amount:       amount,
		LiquidTxHash: liquidTxHash,
	}

	addr := sdk.MustAccAddressFromBech32(cfg.PlanetmintAddress)
	msg := daotypes.NewMsgMintToken(cfg.PlanetmintAddress, &mintRequest)

	_, err = lib.BroadcastTxWithFileLock(addr, msg)
	if err != nil {
		return
	}

	return
}

// Constant rate to be replaced with conversion rate monitor
func (r2p *R2PService) getConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	return rddl * conversionRate
}

func (r2p *R2PService) checkMintRequest(txhash string) (mintRequest *daotypes.QueryGetMintRequestsByHashResponse, err error) {
	cfg := config.GetConfig()
	grcpConn, err := grpc.Dial(
		cfg.PlanetmintRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return mintRequest, err
	}

	daoClient := daotypes.NewQueryClient(grcpConn)
	mintRequest, err = daoClient.GetMintRequestsByHash(
		context.Background(),
		&daotypes.QueryGetMintRequestsByHashRequest{Hash: txhash},
	)

	if strings.Contains(err.Error(), codes.NotFound.String()) {
		return mintRequest, nil
	}

	if err != nil {
		return mintRequest, err
	}

	return
}
