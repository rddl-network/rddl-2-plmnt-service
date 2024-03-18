package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
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

	// store receive address - planetmint address pair
	err = r2p.addConversionRequest(confReceiveAddress, address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storing addresses in DB: " + err.Error()})
		return
	}

	var resBody ReceiveAddressResponse
	resBody.LiquidAddress = confReceiveAddress
	resBody.PlanetmintBeneficiary = address
	c.JSON(http.StatusOK, resBody)
}
