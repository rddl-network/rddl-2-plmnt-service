package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	elements "github.com/rddl-network/elements-rpc"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

type Conversion struct {
	Beneficiary  string `binding:"required" json:"beneficiary"`
	LiquidTXHash string `binding:"required" json:"liquid-tx-hash"`
}

// Request body for REST Endpoint
type MintRequestBody struct {
	Conversion Conversion `binding:"required" json:"conversion"`
	Signature  string     `binding:"required" json:"signature"`
}

func (r2p *R2PService) registerRoutes() {
	r2p.router.POST("/mint", r2p.postMintRequest)
}

// TODO: remove tx hash from url and add test for correctly working signature verification
func (r2p *R2PService) postMintRequest(c *gin.Context) {
	cfg := config.GetConfig()

	// if beneficiary address missing return bad request
	var requestBody MintRequestBody
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check whether mint request already exists
	mr, err := r2p.pmClient.CheckMintRequest(requestBody.Conversion.LiquidTXHash)
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
	url := fmt.Sprintf("http://%s:%s@%s/wallet/%s", cfg.RPCUser, cfg.RPCPass, cfg.RPCHost, cfg.Wallet)
	tx, err := elements.GetTransaction(url, []string{requestBody.Conversion.LiquidTXHash})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while fetching liquid tx: %s", err)})
		return
	}

	// return error if reissuance asset is not in liquid tx
	amt, ok := tx.Amount[cfg.AcceptedAsset]
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("tx does not contain accepted asset: %s", cfg.AcceptedAsset)})
		return
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("accepted asset amount must be positive got: %v", amt)})
		return
	}

	plmntAmount := r2p.getConversion(uint64(amt))
	err = r2p.pmClient.MintPLMNT(requestBody.Conversion.Beneficiary, plmntAmount, requestBody.Conversion.LiquidTXHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error while minting token: %s", err)})
	}
}
