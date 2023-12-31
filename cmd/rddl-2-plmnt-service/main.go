package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	elements "github.com/rddl-network/elements-rpc"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	daotypes "github.com/planetmint/planetmint-go/x/dao/types"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

// Request body for REST Endpoint
type MintRequestBody struct {
	Beneficiary string `json:"beneficiary"`
}

var (
	planetmintAddress string
	rpcHost           string
	rpcUser           string
	rpcPass           string
	pmRPCHost         string
	acceptedAsset     string
	wallet            string
)

func loadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("toml")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		return
	}
	log.Println("no config file found.")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(config.DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	err = configTemplate.Execute(&buffer, config.GetConfig())
	if err != nil {
		return
	}

	err = v.ReadConfig(&buffer)
	if err != nil {
		return
	}
	err = v.SafeWriteConfig()
	if err != nil {
		return
	}

	log.Println("default config file created. please adapt it and restart the application. exiting...")
	os.Exit(0)
	return
}

// Constant rate to be replaced with conversion rate monitor
func getConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	return rddl * conversionRate
}

func checkMintRequest(txhash string) (mintRequest *daotypes.QueryGetMintRequestsByHashResponse, err error) {
	grcpConn, err := grpc.Dial(
		pmRPCHost,
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

func mintPLMNT(beneficiary string, amount uint64, liquidTxHash string) (err error) {
	mintRequest := daotypes.MintRequest{
		Beneficiary:  beneficiary,
		Amount:       amount,
		LiquidTxHash: liquidTxHash,
	}

	addr := sdk.MustAccAddressFromBech32(planetmintAddress)
	msg := daotypes.NewMsgMintToken(planetmintAddress, &mintRequest)

	_, err = lib.BroadcastTxWithFileLock(addr, msg)
	if err != nil {
		return
	}

	return
}

func postMintRequest(c *gin.Context) {
	txhash := c.Param("txhash")

	// if beneficiary address missing return bad request
	var requestBody MintRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		return
	}

	// check whether mint request already exists
	mr, err := checkMintRequest(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while fetching mint request: %s", err)})
		return
	}

	// return because mint request for txhash is already
	if mr != nil {
		c.JSON(http.StatusConflict, gin.H{"msg": "already minted"})
		return
	}

	// fetch liquid tx for amount of rddl
	url := fmt.Sprintf("http://%s:%s@%s/wallet/%s", rpcUser, rpcPass, rpcHost, wallet)
	tx, err := elements.GetTransaction(url, `"`+txhash+`"`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while fetching liquid tx: %s", err)})
		return
	}

	// return error if reissuance asset is not in liquid tx
	amt, ok := tx.Amount[acceptedAsset]
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("tx does not contain accepted asset: %s", acceptedAsset)})
		return
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("accepted asset amount must be positive got: %v", amt)})
		return
	}

	plmntAmount := getConversion(uint64(amt))
	err = mintPLMNT(requestBody.Beneficiary, plmntAmount, txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while minting token: %s", err)})
	}
}

func startWebService(config *viper.Viper) {
	router := gin.Default()
	router.POST("/mint/:txhash", postMintRequest)

	bindAddress := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	_ = router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

var libConfig *lib.Config

func main() {
	config, err := loadConfig("./")
	if err != nil {
		log.Fatalf("fatal error loading config file: %s", err)
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
		log.Fatalf("chain id must not be empty")
	}
	libConfig.SetChainID(planetmintChainID)

	acceptedAsset = config.GetString("accepted-asset")
	wallet = config.GetString("wallet")
	if acceptedAsset == "" || wallet == "" {
		panic("Could not read configuration")
	}

	startWebService(config)
}
