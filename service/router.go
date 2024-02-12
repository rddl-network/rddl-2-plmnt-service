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

	// check whether mint request already exists
	mr, err := r2p.pmClient.CheckMintRequest(requestBody.Conversion.LiquidTxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching mint request: " + err.Error()})
		return
	}

	// return because mint request for txhash is already
	if mr != nil {
		c.JSON(http.StatusConflict, gin.H{"msg": "already minted"})
		return
	}

	// fetch liquid tx for amount of rddl
	url := fmt.Sprintf("http://%s:%s@%s/wallet/%s", cfg.RPCUser, cfg.RPCPass, cfg.RPCHost, cfg.Wallet)
	tx, err := r2p.eClient.GetTransaction(url, []string{requestBody.Conversion.LiquidTxHash})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching liquid tx: " + err.Error()})
		return
	}

	addresses, err := r2p.eClient.DeriveAddresses(url, []string{requestBody.Conversion.Descriptor})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while deriving liquid addresses: " + err.Error()})
		return
	}

	if !slices.Contains(addresses, tx.Details[0].Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction details do not contain address derived from descriptor: " + requestBody.Conversion.Descriptor})
		return
	}

	// return error if reissuance asset is not in liquid tx
	amt, ok := tx.Amount[cfg.AcceptedAsset]
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "tx does not contain accepted asset: " + cfg.AcceptedAsset})
		return
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("accepted asset amount must be positive got: %v", amt)})
		return
	}

	plmntAmount := r2p.getConversion(uint64(amt))
	err = r2p.pmClient.MintPLMNT(requestBody.Conversion.Beneficiary, plmntAmount, requestBody.Conversion.LiquidTxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while minting token: " + err.Error()})
	}
}
