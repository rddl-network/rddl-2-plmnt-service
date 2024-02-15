package service

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

type Conversion struct {
	Beneficiary  string `binding:"required" json:"beneficiary"`
	LiquidTxHash string `binding:"required" json:"liquid-tx-hash"`
	Descriptor   string `binding:"required" json:"descriptor"`
}

// Request body for REST Endpoint
type MintRequestBody struct {
	Conversion Conversion `binding:"required" json:"conversion"`
	Signature  string     `binding:"required" json:"signature"`
}

func (r2p *R2PService) configureRouter() {
	r2p.router.Use(gin.Logger())
	r2p.router.Use(gin.Recovery())
}

func (r2p *R2PService) registerRoutes() {
	r2p.router.POST("/mint", r2p.postMintRequest)
}

func (r2p *R2PService) postMintRequest(c *gin.Context) {
	cfg := config.GetConfig()

	// if beneficiary address missing return bad request
	var requestBody MintRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := r2p.VerifyMessage(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while verifying signature: " + err.Error()})
		return
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	// check if mint request has already been issued
	code, err := r2p.checkMintRequest(requestBody.Conversion.LiquidTxHash)
	if err != nil {
		c.JSON(code, gin.H{"error": "error while checking mint request: " + err.Error()})
		return
	}

	// fetch liquid tx for amount of rddl
	url := fmt.Sprintf("http://%s:%s@%s/wallet/%s", cfg.RPCUser, cfg.RPCPass, cfg.RPCHost, cfg.Wallet)
	tx, err := r2p.eClient.GetTransaction(url, []string{requestBody.Conversion.LiquidTxHash})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching liquid tx: " + err.Error()})
		return
	}

	// check if provided address and descriptor are part of the transaction
	code, err = r2p.checkAddress(url, tx.Details[0].Address, requestBody.Conversion.Descriptor)
	if err != nil {
		c.JSON(code, gin.H{"error": "error while checking address descriptor: " + err.Error()})
		return
	}

	// check if asset is in liquid tx
	amt, err := r2p.checkAsset(tx.Amount, cfg.AcceptedAsset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking asset: " + err.Error()})
		return
	}

	plmntAmount := r2p.getConversion(uint64(amt))
	err = r2p.pmClient.MintPLMNT(requestBody.Conversion.Beneficiary, plmntAmount, requestBody.Conversion.LiquidTxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while minting token: " + err.Error()})
	}
}

func (r2p *R2PService) checkMintRequest(liquidTxHash string) (code int, err error) {
	// check whether mint request already exists
	mr, err := r2p.pmClient.CheckMintRequest(liquidTxHash)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error while fetching mint request: %v", err)
	}

	// return because mint request for txhash is already
	if mr != nil {
		return http.StatusConflict, fmt.Errorf("already minted")
	}
	return
}

func (r2p *R2PService) checkAddress(url string, txAddress string, descriptor string) (code int, err error) {
	addresses, err := r2p.eClient.DeriveAddresses(url, []string{descriptor})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error while deriving liquid addresses: %v", err)
	}

	if !slices.Contains(addresses, txAddress) {
		return http.StatusBadRequest, fmt.Errorf("transaction details do not contain address derived from descriptor: %s", descriptor)
	}
	return
}

func (r2p *R2PService) checkAsset(amounts map[string]float64, asset string) (amount float64, err error) {
	// return error if reissuance asset is not in liquid tx
	amt, ok := amounts[asset]
	if !ok {
		return 0, fmt.Errorf("tx does not contain accepted asset: " + asset)
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		return 0, fmt.Errorf("accepted asset amount must be positive got: %v", amt)
	}
	return amt, nil
}
