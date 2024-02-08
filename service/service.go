package service

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type R2PService struct {
	router   *gin.Engine
	pmClient IPlanetmintClient
	eClient  IElementsClient
}

func NewR2PService(router *gin.Engine, pmClient IPlanetmintClient, eClient IElementsClient) *R2PService {
	service := &R2PService{router: router, pmClient: pmClient, eClient: eClient}
	service.configureRouter()
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

func (r2p *R2PService) ValidateConversionSignature(conversion Conversion, signature string, publicKey string) (isValid bool, err error) {
	conversionBytes, err := json.Marshal(conversion)
	if err != nil {
		return false, err
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}

	pubKey := &secp256k1.PubKey{Key: publicKeyBytes}

	isValid = pubKey.VerifySignature(conversionBytes, signatureBytes)
	if !isValid {
		return false, errors.New("invalid signature")
	}

	return
}
