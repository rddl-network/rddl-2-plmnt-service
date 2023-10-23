package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	elements "github.com/rddl-network/elements-rpc"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type CheckMintRequestResponse struct {
	Request CheckMintRequest `json:"mintRequest"`
}

type CheckMintRequest struct {
	Beneficiary  string `json:"beneficiary"`
	Amount       string `json:"amount"`
	LiquidTXHash string `json:"liquidTXHash"`
}

type MintRequest struct {
	Beneficiary  string `json:"beneficiary"`
	Amount       uint64 `json:"amount"`
	LiquidTXHash string `json:"liquidTXHash"`
}

type MintRequestBody struct {
	Beneficiary string `json:"beneficiary"`
}

var (
	planetmint        string
	planetmintAddress string
	planetmintKeyring string
	rpcHost           string
	rpcUser           string
	rpcPass           string
	reissuanceAsset   string
)

func loadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return
	}

	planetmint = v.GetString("PLANETMINT_GO")
	planetmintAddress = v.GetString("PLANETMINT_ADDRESS")
	planetmintKeyring = v.GetString("PLANETMINT_KEYRING")
	if err != nil || planetmint == "" || planetmintAddress == "" {
		panic("Could not read configuration")
	}

	rpcHost = v.GetString("RPC_HOST")
	rpcUser = v.GetString("RPC_USER")
	rpcPass = v.GetString("RPC_PASS")
	if rpcHost == "" || rpcUser == "" || rpcPass == "" {
		panic("Could not read configuration")
	}

	reissuanceAsset = v.GetString("REISSUANCE_ASSET")
	if reissuanceAsset == "" {
		panic("Could not read configuration")
	}

	return
}

// Constant rate to be replaced with conversion rate monitor
func getConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	return rddl * conversionRate
}

func checkMintRequest(txhash string) (mintRequest CheckMintRequestResponse, err error) {
	cmdStr := fmt.Sprintf("%s query dao get-mint-requests-by-hash %s -o json", planetmint, txhash)

	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Run()
	errStr := stderr.String()
	if strings.Contains(errStr, "mint request not found") {
		return mintRequest, err
	}

	err = json.Unmarshal(stdout.Bytes(), &mintRequest)
	if err != nil {
		return mintRequest, err
	}

	return
}

func mintPLMNT(beneficiary string, amount uint64, liquidTxHash string) (err error) {
	mintRequest := MintRequest{
		Beneficiary:  beneficiary,
		Amount:       amount,
		LiquidTXHash: liquidTxHash,
	}

	mrJSON, err := json.Marshal(mintRequest)
	if err != nil {
		return err
	}

	cmdStr := fmt.Sprintf("%s tx dao mint-token '%s' --from %s -y", planetmint, string(mrJSON), planetmintAddress)

	if planetmintKeyring != "" {
		cmdStr = fmt.Sprintf("%s --keyring-backend %s", cmdStr, planetmintKeyring)
	}

	cmd := exec.Command("bash", "-c", cmdStr)

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
	if mr.Request.Beneficiary != "" {
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
	if _, ok := tx.Amount[reissuanceAsset]; !ok {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("tx does not contain reissuance asset: %s", reissuanceAsset)})
		return
	}

	plmntAmount := getConversion(uint64(tx.Amount[reissuanceAsset]))
	err = mintPLMNT(requestBody.Beneficiary, plmntAmount, txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while minting token: %s", err)})
	}
}

func startWebService(config *viper.Viper) {
	router := gin.Default()
	router.POST("/mint/:txhash", postIssue)

	bindAddress := config.GetString("SERVICE_BIND")
	servicePort := config.GetString("SERVICE_PORT")
	router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

func main() {
	config, err := loadConfig("./")
	if err != nil {
		panic(err)
	}

	startWebService(config)
}
