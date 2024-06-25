package main

import (
	stdlog "log"

	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	log "github.com/rddl-network/go-logger"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
)

var (
	planetmintAddress string
	rpcHost           string
	rpcUser           string
	rpcPass           string
	pmRPCHost         string
	acceptedAsset     string
	wallet            string
)

var libConfig *lib.Config

func main() {
	config, err := config.LoadConfig("./")
	if err != nil {
		stdlog.Fatalf("fatal error loading config file: %s", err)
	}

	planetmintAddress = config.GetString("planetmint-address")
	if err != nil || planetmintAddress == "" {
		panic("Could not read configuration")
	}

	rpcHost = config.GetString("rpc-host")
	rpcUser = config.GetString("rpc-user")
	rpcPass = config.GetString("rpc-pass")
	pmRPCHost = config.GetString("planetmint-rpc-host")
	if rpcHost == "" || rpcUser == "" || rpcPass == "" || pmRPCHost == "" {
		panic("Could not read configuration")
	}

	encodingConfig := app.MakeEncodingConfig()

	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)

	planetmintChainID := config.GetString("planetmint-chain-id")
	if planetmintChainID == "" {
		stdlog.Fatalf("chain id must not be empty")
	}
	libConfig.SetChainID(planetmintChainID)

	acceptedAsset = config.GetString("accepted-asset")
	wallet = config.GetString("wallet")
	if acceptedAsset == "" || wallet == "" {
		panic("Could not read configuration")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	db, err := leveldb.OpenFile("./conversions.db", nil)
	if err != nil {
		db.Close()
		stdlog.Fatal(err)
	}
	defer db.Close()
	pmClient := service.NewPlanetmintClient()
	eClient := service.NewElementsClient()
	logger := log.GetLogger(config.GetString("log-level"))
	service := service.NewR2PService(router, pmClient, eClient, db, logger)

	if err = service.Run(config); err != nil {
		stdlog.Panicf("error occurred while spinning up service: %v", err)
	}
}
