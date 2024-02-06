package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type R2PService struct {
	router   *gin.Engine
	pmClient IPlanetmintClient
}

func NewR2PService(router *gin.Engine, pmClient IPlanetmintClient) *R2PService {
	service := &R2PService{router: router, pmClient: pmClient}
	service.registerRoutes()
	return service
}

func (r2p *R2PService) Run(config *viper.Viper) {
	bindAddress := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	_ = r2p.router.Run(fmt.Sprintf("%s:%s", bindAddress, servicePort))
}

// Constant rate to be replaced with conversion rate monitor
func (r2p *R2PService) getConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	return rddl * conversionRate
}
