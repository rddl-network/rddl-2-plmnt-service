package service

import (
	"errors"
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
	r2p.router.GET("/receiveaddress/:plmntaddress", r2p.getReceiveAddress)
}

type AddressBody struct {
	LiquidAddress         string `binding:"required" json:"liquid-address"`
	PlanetmintBeneficiary string `binding:"required" json:"planetmint-beneficiary"`
}

func (r2p *R2PService) getReceiveAddress(c *gin.Context) {
	cfg := config.GetConfig()
	address := c.Param("plmntaddress")

	// is legit machine address?
	resp, err := r2p.pmClient.IsLegitMachine(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if resp.GetMachine().Address != address {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "different machine resolved: " + resp.GetMachine().Address + " instead of " + address})
		return
	}

	// derive new receive address
	confReceiveAddress, err := r2p.eClient.GetNewAddress(cfg.GetElementsURL(), []string{
		``,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "getting new receive address: " + err.Error()})
		return
	}
	addressInfo, err := r2p.eClient.GetAddressInfo(cfg.GetElementsURL(), []string{confReceiveAddress})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "getting receive address information: " + err.Error()})
		return
	}

	// store receive address - planetmint address pair
	err = r2p.addConversionRequest(confReceiveAddress, addressInfo.Unconfidential, address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storing addresses in DB: " + err.Error()})
		return
	}

	var resBody AddressBody
	resBody.LiquidAddress = confReceiveAddress
	resBody.PlanetmintBeneficiary = address
	c.JSON(http.StatusOK, resBody)
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
	tx, err := r2p.eClient.GetTransaction(cfg.GetElementsURL(), []string{requestBody.Conversion.LiquidTxHash})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching liquid tx: " + err.Error()})
		return
	}

	// check if enough confirmations
	if tx.Confirmations < cfg.Confirmations {
		c.JSON(http.StatusConflict, gin.H{"error": "not enough confirmations on tx"})
		return
	}

	// check if provided address and descriptor are part of the transaction
	code, err = r2p.checkAddress(cfg.GetElementsURL(), tx.Details[0].Address, requestBody.Conversion.Descriptor)
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

func (r2p *R2PService) checkAddress(url string, txAddress string, descriptor string) (code int, err error) {
	addresses, err := r2p.eClient.DeriveAddresses(url, []string{descriptor})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error while deriving liquid addresses: %w", err)
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
		return 0, errors.New("tx does not contain accepted asset: " + asset)
	}

	// check if amount is positive otherwise return error
	if amt <= 0 {
		return 0, fmt.Errorf("accepted asset amount must be positive got: %v", amt)
	}
	return amt, nil
}
