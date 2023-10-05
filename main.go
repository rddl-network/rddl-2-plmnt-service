package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type MintRequest struct {
	Beneficiary  string
	Amount       uint64
	LiquidTXHash string
}

var planetmint string
var planetmintAddress string
var planetmintKeyring string
var client *rpcclient.Client

func loadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return
		// panic(err)
	}

	return
}

func mintPLMNT(beneficiary string, amount uint64, liquidTxHash string) {
	mintRequest := MintRequest{
		Beneficiary:  beneficiary,
		Amount:       amount,
		LiquidTXHash: liquidTxHash,
	}

	mrJSON, err := json.Marshal(mintRequest)
	if err != nil {
		return
	}

	cmdStr := fmt.Sprintf("%s tx dao mint-token %s --from %s -y", planetmint, string(mrJSON), planetmintAddress)

	if planetmintKeyring != "" {
		cmdStr = fmt.Sprintf("%s --keyring-backend %s", cmdStr, planetmintKeyring)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	fmt.Println("Output: ", string(out))
}

func postIssue(c *gin.Context) {
	txhash := c.Param("txhash")

	chainhash, err := chainhash.NewHashFromStr(txhash)
	if err != nil {
		fmt.Println(err)
		return
	}

	txResult, err := client.GetTransaction(chainhash)
	if err != nil {
		fmt.Println(err)
		return
	}

	// TODO: read txResult beneficiary/amount
}

func startWebService(config *viper.Viper) {
	router := gin.Default()
	router.POST("/issue/:txhash", postIssue)

	bindAddress := config.GetString("SERVICE_BIND")
	servicePort := config.GetString("SERVICE_PORT")
	router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

func main() {
	config, err := loadConfig("./")

	connCfg := &rpcclient.ConnConfig{
		Host:         "planetmint-go-testnet-explorer.rddl.io:18884",
		User:         "",
		Pass:         "",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)
}
