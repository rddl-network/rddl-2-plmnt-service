package service_test

import (
	"log"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/rddl-network/rddl-2-plmnt-service/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestPeriodicCheck(t *testing.T) {
	t.Parallel()
	_, err := config.LoadConfig("./")
	assert.NoError(t, err)
	router := gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock := testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.OpenFile("./conversions.db", nil)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	r2p := service.NewR2PService(router, pmClientMock, eClientMock, db)

	eClientMock.EXPECT().ListReceivedByAddress(gomock.Any(), gomock.Any()).Return(testutil.ReceivedTxByAddressArray1Tx, nil).AnyTimes()
	pmClientMock.EXPECT().CheckMintRequest(gomock.Any()).Return(nil, nil).AnyTimes()
	pmClientMock.EXPECT().MintPLMNT(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	var conversion service.ConversionRequest
	conversion.ConfidentialAddress = "tlq1qqfz5fmd860877mm7ka7s5a3ryzeajd7xsamedk4cljtlla7tpzx3zux9sk6msuth78rtk7u4whn2nkxe8l9uyy9pcd9semy9m"
	_, err = r2p.ExecutePotentialConversion(conversion)
	assert.NoError(t, err)
}
