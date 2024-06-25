package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/rddl-network/rddl-2-plmnt-service/types"
)

func (r2p *R2PService) configureRouter() {
	r2p.router.Use(gin.Logger())
	r2p.router.Use(gin.Recovery())
}

func (r2p *R2PService) registerRoutes() {
	r2p.router.GET("/receiveaddress/:plmntaddress", r2p.getReceiveAddress)
}

func (r2p *R2PService) getReceiveAddress(c *gin.Context) {
	cfg := config.GetConfig()
	address := c.Param("plmntaddress")

	// is legit planetmint address?
	valid, err := VerifyAddress(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid planetmint address"})
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

	// store receive address - planetmint address pair
	err = r2p.addConversionRequest(confReceiveAddress, address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storing addresses in DB: " + err.Error()})
		return
	}

	var resBody types.ReceiveAddressResponse
	resBody.LiquidAddress = confReceiveAddress
	resBody.PlanetmintBeneficiary = address
	c.JSON(http.StatusOK, resBody)
}
