package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

type R2PService struct {
	router     *gin.Engine
	pmClient   IPlanetmintClient
	eClient    IElementsClient
	db         *leveldb.DB
	dbMutex    sync.Mutex // Mutex to synchronize write operations
	tickerList []*time.Ticker
}

func NewR2PService(router *gin.Engine, pmClient IPlanetmintClient, eClient IElementsClient, db *leveldb.DB) *R2PService {
	service := &R2PService{router: router, pmClient: pmClient, eClient: eClient, db: db}
	gin.SetMode(gin.ReleaseMode)
	service.configureRouter()
	service.registerRoutes()
	service.registerPeriodicTasks()
	return service
}

func (r2p *R2PService) Run(config *viper.Viper) (err error) {
	serviceBind := config.GetString("service-bind")
	servicePort := config.GetString("service-port")
	return r2p.router.Run(fmt.Sprintf("%s:%s", serviceBind, servicePort))
}

// Constant rate to be replaced with conversion rate monitor
func GetConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100)
	plmntAmount := rddl * conversionRate
	plmnt = plmntAmount / 100000000 // TODO: to be replaced by util.Factor
	return
}

// Constant rate to be replaced with conversion rate monitor
func ShiftValue(plmnt uint64, digitsToSkip int) uint64 {
	plmntShifted := plmnt >> digitsToSkip
	return plmntShifted
}
