package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/rddl-network/rddl-2-plmnt-service/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestGetReceiveAddressRoute(t *testing.T) {
	router := gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock := testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.OpenFile("./conversions.db", nil)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}
	// defer db.Close()
	_ = service.NewR2PService(router, pmClientMock, eClientMock, db)

	eClientMock.EXPECT().GetNewAddress(gomock.Any(), gomock.Any()).Return(testutil.ConfidentialAddr, nil).AnyTimes()

	tests := []struct {
		desc              string
		planetmintAddress string
		resBody           service.ReceiveAddressResponse
		code              int
		errorMsg          string
	}{
		{
			desc:              "valid request",
			planetmintAddress: testutil.PlanetmintAddress,
			resBody: service.ReceiveAddressResponse{
				LiquidAddress:         testutil.ConfidentialAddr,
				PlanetmintBeneficiary: testutil.PlanetmintAddress,
			},
			code:     200,
			errorMsg: "",
		},
		{
			desc:              "missing request fields",
			planetmintAddress: "",
			resBody:           service.ReceiveAddressResponse{},
			code:              404,
			errorMsg:          "404 page not found",
		},
		{
			desc:              "Invalid planetmint machine address",
			planetmintAddress: "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx",
			resBody:           service.ReceiveAddressResponse{},
			code:              400,
			errorMsg:          "{\"error:\":\"different machine resolved: plmnt1683t0us0r85840nsepx6jrk2kjxw7zrcnkf0rp instead of plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx\"}",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/receiveaddress/"+tc.planetmintAddress, bytes.NewBuffer([]byte{}))
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
			if w.Code != 200 {
				assert.Equal(t, tc.errorMsg, w.Body.String())
			} else {
				var result service.ReceiveAddressResponse
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
