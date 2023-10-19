package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/btcsuite/btcd/rpcclient"
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

type GetTransactionDetailsResult struct {
	Account           string   `json:"account"`
	Address           string   `json:"address,omitempty"`
	Amount            float64  `json:"amount"`
	Category          string   `json:"category"`
	InvolvesWatchOnly bool     `json:"involveswatchonly,omitempty"`
	Fee               *float64 `json:"fee,omitempty"`
	Vout              uint32   `json:"vout"`
}

// GetTransactionResult models the data from the gettransaction command.
type GetTransactionResult struct {
	Amount          map[string]float64            `json:"amount"`
	Fee             float64                       `json:"fee,omitempty"`
	Confirmations   int64                         `json:"confirmations"`
	BlockHash       string                        `json:"blockhash"`
	BlockIndex      int64                         `json:"blockindex"`
	BlockTime       int64                         `json:"blocktime"`
	TxID            string                        `json:"txid"`
	WalletConflicts []string                      `json:"walletconflicts"`
	Time            int64                         `json:"time"`
	TimeReceived    int64                         `json:"timereceived"`
	Details         []GetTransactionDetailsResult `json:"details"`
	Hex             string                        `json:"hex"`
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
	client            *rpcclient.Client
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

func getLiquidTx(txhash string) (liquidTx GetTransactionResult, err error) {
	cmdStr := fmt.Sprintf("elements-cli -rpcpassword=%s -rpcuser=%s -rpcport=18884 -rpcconnect=%s gettransaction %s", rpcPass, rpcUser, rpcHost, txhash)
	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		fmt.Println("could not run command: ", err)
		return liquidTx, err
	}

	err = json.Unmarshal(stdout.Bytes(), &liquidTx)
	if err != nil {
		return liquidTx, err
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
	tx, err := getLiquidTx(txhash)
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

func setupRPCClient(config *viper.Viper) *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         config.GetString("RPC_HOST"),
		User:         config.GetString("RPC_USER"),
		Pass:         config.GetString("RPC_PASS"),
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
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

	client = setupRPCClient(config)
	defer client.Shutdown()

	startWebService(config)
}
