package service_test

import (
	stdlog "log"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/planetmint/planetmint-go/util"
	log "github.com/rddl-network/go-utils/logger"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/rddl-network/rddl-2-plmnt-service/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestPeriodicCheck(t *testing.T) {
	_ = config.GetConfig()

	router := gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock := testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		db.Close()
		stdlog.Fatal(err)
	}
	defer db.Close()
	r2p := service.NewR2PService(router, pmClientMock, eClientMock, db, log.GetLogger(log.DEBUG))

	eClientMock.EXPECT().ListReceivedByAddress(gomock.Any(), gomock.Any()).Return(testutil.ReceivedTxByAddressArray1Tx, nil).AnyTimes()
	pmClientMock.EXPECT().CheckMintRequest(gomock.Any()).Return(nil, nil).AnyTimes()
	pmClientMock.EXPECT().MintPLMNT(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	var conversion service.ConversionRequest
	conversion.ConfidentialAddress = "tlq1qqfz5fmd860877mm7ka7s5a3ryzeajd7xsamedk4cljtlla7tpzx3zux9sk6msuth78rtk7u4whn2nkxe8l9uyy9pcd9semy9m"
	_, err = r2p.ExecutePotentialConversion(conversion)
	assert.NoError(t, err)
}

func TestConversion(t *testing.T) {
	convertedAmount := util.RDDLToken2Uint(570330.47944743)
	plmntAmount := service.GetConversion(convertedAmount)
	assert.Equal(t, uint64(57033047), plmntAmount)
}
