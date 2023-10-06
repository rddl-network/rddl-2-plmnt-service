package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type MintRequestResponse struct {
	Request MintRequest `json:"mintRequest"`
}

type MintRequest struct {
	Beneficiary  string `json:"beneficiary"`
	Amount       string `json:"amount"`
	LiquidTXHash string `json:"liquidTXHash"`
}

// TODO: finalize details to resemble liquid
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

var (
	planetmint        string
	planetmintAddress string
	planetmintKeyring string
	rpcHost           string
	rpcUser           string
	rpcPass           string
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

	return
}

// TODO: how to check response in case of rpc error
func checkMintRequest(txhash string) (mintRequest MintRequestResponse, err error) {
	cmdStr := fmt.Sprintf("%s query dao get-mint-requests-by-hash %s -o json", planetmint, txhash)

	cmd := exec.Command("bash", "-c", cmdStr)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
		return mintRequest, err
	}
	fmt.Println("Output: ", string(out))

	err = json.Unmarshal(out, &mintRequest)
	if err != nil {
		return mintRequest, err
	}

	return
}

// TODO: amount should be uint64
func mintPLMNT(beneficiary string, amount string, liquidTxHash string) {
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

func getLiquidTx(txhash string) (liquidTx GetTransactionResult, err error) {
	cmdStr := fmt.Sprintf("elements-cli -rpcpassword=%s -rpcuser=%s -rpcport=18884 -rpcconnect=%s gettransaction %s", rpcPass, rpcUser, rpcHost, txhash)
	cmd := exec.Command("bash", "-c", cmdStr)

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("could not run command: ", err)
	}
	fmt.Println("Output: ", string(out))

	err = json.Unmarshal(out, &liquidTx)
	if err != nil {
		return liquidTx, err
	}

	return
}

func postIssue(c *gin.Context) {
	txhash := c.Param("txhash")

	fmt.Println("CHECK MINT REQUEST")
	mr, err := checkMintRequest(txhash)
	if err != nil {
		fmt.Println(err)
		// no return becouse error also means not found
	}

	// return tx containing msg if error
	fmt.Println("MintRequest: ", mr.Request)

	fmt.Println("GET LIQUID HASH")
	tx, err := getLiquidTx(txhash)
	if err != nil {
		fmt.Println(err)
		return
	}

	mintPLMNT("bene", fmt.Sprintf("%f", tx.Amount["rddl"]*100), txhash)
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
