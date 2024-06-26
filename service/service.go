package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/rddl-network/go-logger"
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
	logger     log.AppLogger
}

func NewR2PService(router *gin.Engine, pmClient IPlanetmintClient, eClient IElementsClient, db *leveldb.DB, logger log.AppLogger) *R2PService {
	service := &R2PService{router: router, pmClient: pmClient, eClient: eClient, db: db, logger: logger}
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

// TODO: Constant rate to be replaced with conversion rate monitor
// Cut away the PLMNT fractions as planetmint only works with natural numbers
func GetConversion(rddl uint64) (plmnt uint64) {
	conversionRate := uint64(100) // apply conversion rate
	plmntAmount := rddl * conversionRate
	// TODO: 100000000 to be replaced by util.Factor
	plmnt = plmntAmount / 100000000

	return
}
