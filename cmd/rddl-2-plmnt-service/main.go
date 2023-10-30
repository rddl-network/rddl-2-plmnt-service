package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/codec"
	elements "github.com/rddl-network/elements-rpc"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	daotypes "github.com/planetmint/planetmint-go/x/dao/types"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

// Request body for REST Endpoint
type MintRequestBody struct {
	Beneficiary string `json:"beneficiary"`
}

var (
	planetmint        string
	planetmintAddress string
	rpcHost           string
	rpcUser           string
	rpcPass           string
	pmRPCHost         string
	reissuanceAsset   string
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
	if err == nil {
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

	mrJSON, err := json.Marshal(mintRequest)
	if err != nil {
		return err
	}

	cmd := exec.Command(planetmint, "tx", "dao", "mint-token", string(mrJSON), "--from", planetmintAddress)

	err = cmd.Run()
	if err != nil {
		fmt.Println("could not run command: ", err)
		return err
	}

	return
}

func postIssue(c *gin.Context) {
	txhash := c.Param("txhash")

	// if beneficiary address missing return bad request
	var requestBody MintRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		return
	}

	// check if mint request is already existant
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
	url := fmt.Sprintf("http://%s:%s@%s", rpcUser, rpcPass, rpcHost)
	tx, err := elements.GetWalletTx(url, txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while fetching liquid tx: %s", err)})
		return
	}

	// return error if reissuance asset is not in liquid tx
	amt, ok := tx.Amount[reissuanceAsset]
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("tx does not contain reissuance asset: %s", reissuanceAsset)})
		return
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("reissaunce asset amount must be positive got: %v", amt)})
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
	router.POST("/mint/:txhash", postIssue)

	bindAddress := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	_ = router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

func main() {
	config, err := loadConfig("../../")
	if err != nil {
		log.Fatalf("fatal error loading config file: %s", err)
	}

	planetmint = config.GetString("planetmint")
	planetmintAddress = config.GetString("planetmint-address")
	if err != nil || planetmint == "" || planetmintAddress == "" {
		panic("Could not read configuration")
	}

	rpcHost = config.GetString("rpc-host")
	rpcUser = config.GetString("rpc-user")
	rpcPass = config.GetString("rpc-pass")
	pmRPCHost = config.GetString("planetmint-rpc-host")
	if rpcHost == "" || rpcUser == "" || rpcPass == "" || pmRPCHost == "" {
		panic("Could not read configuration")
	}

	reissuanceAsset = config.GetString("reissuance-asset")
	if reissuanceAsset == "" {
		panic("Could not read configuration")
	}

	startWebService(config)
}
