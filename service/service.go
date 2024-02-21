package service

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

// Text used to signify that a signed message follows and to prevent
// inadvertently signing a transaction.
const messageSignatureHeader = "Bitcoin Signed Message:\n"

var ErrInvalidDescriptor = errors.New("invalid descriptor: input is malformed")

type R2PService struct {
	router   *gin.Engine
	pmClient IPlanetmintClient
	eClient  IElementsClient
	db       *leveldb.DB
}

func NewR2PService(router *gin.Engine, pmClient IPlanetmintClient, eClient IElementsClient, db *leveldb.DB) *R2PService {
	service := &R2PService{router: router, pmClient: pmClient, eClient: eClient, db: db}
	service.configureRouter()
	service.registerRoutes()
	return service
}

func (r2p *R2PService) Run(config *viper.Viper) (err error) {
	serviceBind := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	return r2p.router.Run(fmt.Sprintf("%s:%s", serviceBind, servicePort))
}

// Constant rate to be replaced with conversion rate monitor
func (r2p *R2PService) getConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	return rddl * conversionRate
}

func (r2p *R2PService) VerifyMessage(conversionWithSignature MintRequestBody) (valid bool, err error) {
	re := regexp.MustCompile(`wpkh\(\[.*\](.*)\)#.*`)
	match := re.FindStringSubmatch(conversionWithSignature.Conversion.Descriptor)
	if len(match) < 2 {
		err = ErrInvalidDescriptor
		return
	}
	conversionPK, err := hex.DecodeString(match[1])
	if err != nil {
		return
	}

	msg, err := json.Marshal(conversionWithSignature.Conversion)
	if err != nil {
		return
	}
	message := string(msg)

	sig, err := base64.StdEncoding.DecodeString(conversionWithSignature.Signature)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = wire.WriteVarString(&buf, 0, messageSignatureHeader)
	if err != nil {
		panic(err)
	}
	err = wire.WriteVarString(&buf, 0, message)
	if err != nil {
		panic(err)
	}
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, wasCompressed, err := btcecdsa.RecoverCompact(sig, expectedMessageHash)
	if err != nil {
		panic(err)
	}

	// Reconstruct the pubkey hash.
	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}
	valid = bytes.Equal(serializedPK, conversionPK)

	return
}
