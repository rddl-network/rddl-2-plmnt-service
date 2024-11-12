package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	log "github.com/rddl-network/go-utils/logger"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/rddl-network/rddl-2-plmnt-service/testutil"
	"github.com/rddl-network/rddl-2-plmnt-service/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestGetReceiveAddressRoute(t *testing.T) {
	router := gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock := testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		db.Close()
		stdlog.Fatal(err)
	}
	_ = service.NewR2PService(router, pmClientMock, eClientMock, db, log.GetLogger(log.DEBUG))

	eClientMock.EXPECT().GetNewAddress(gomock.Any(), gomock.Any()).Return(testutil.ConfidentialAddr, nil).AnyTimes()

	tests := []struct {
		desc              string
		planetmintAddress string
		resBody           types.ReceiveAddressResponse
		code              int
		errorMsg          string
	}{
		{
			desc:              "valid request",
			planetmintAddress: testutil.PlanetmintAddress,
			resBody: types.ReceiveAddressResponse{
				LiquidAddress:         testutil.ConfidentialAddr,
				PlanetmintBeneficiary: testutil.PlanetmintAddress,
			},
			code:     200,
			errorMsg: "",
		},
		{
			desc:              "missing request fields",
			planetmintAddress: "",
			resBody:           types.ReceiveAddressResponse{},
			code:              404,
			errorMsg:          "404 page not found",
		},
		{
			desc:              "Invalid planetmint address",
			planetmintAddress: "plmnt1w5dww355ck3u4w4hjxcx",
			resBody:           types.ReceiveAddressResponse{},
			code:              400,
			errorMsg:          "{\"error\":\"decoding bech32 failed: invalid checksum (expected j98ean got 4hjxcx)\"}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/receiveaddress/"+tc.planetmintAddress, bytes.NewBuffer([]byte{}))
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
			if w.Code != 200 {
				assert.Equal(t, tc.errorMsg, w.Body.String())
			} else {
				var result types.ReceiveAddressResponse
				err = json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tc.resBody, result)
			}
		})
	}
}

func TestPlmntAddress(t *testing.T) {
	for _, testcase := range []struct {
		address string
		valid   bool
	}{
		{"plmnt000000000000000000000000000000000000000000000000000000000000", false},
		{"plmnt10mq5nj8jhh27z7ejnz2ql3nh0qhzjnfvy50877", true},
		{"plmnt10mq5nj8jhh27z7ejnz2ql3nh0qhzjnfvy5877", false},
		{"plmnt10mq5nj8jhh27z7ejnz2ql3nh0qhzjnfvyx5877", false},
		{"cosmos140e7u946a2nqqkvcnjpjm83d0ynsqem8g840tx", false},
	} {
		valid, err := service.VerifyAddress(testcase.address)
		if testcase.valid {
			assert.NoError(t, err)
			assert.True(t, valid)
		} else {
			assert.False(t, valid)
		}
	}
}
