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

func setupService(t *testing.T) (app *service.R2PService, router *gin.Engine, pmClientMock *testutil.MockIPlanetmintClient, eClientMock *testutil.MockIElementsClient, db *leveldb.DB) {
	router = gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock = testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock = testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.OpenFile("./conversions.db", nil)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}
	defer db.Close()
	app = service.NewR2PService(router, pmClientMock, eClientMock, db)
	return
}

func TestPostMintRequestRoute(t *testing.T) {
	t.Parallel()
	_, router, pmClientMock, eClientMock, _ := setupService(t)

	pmClientMock.EXPECT().CheckMintRequest(gomock.Any()).Return(nil, nil).AnyTimes()
	pmClientMock.EXPECT().MintPLMNT(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	eClientMock.EXPECT().GetTransaction(gomock.Any(), gomock.Any()).Return(testutil.GetTransactionResult, nil).AnyTimes()
	eClientMock.EXPECT().DeriveAddresses(gomock.Any(), gomock.Any()).Return(testutil.DeriveAddressesResult, nil).AnyTimes()

	tests := []struct {
		desc    string
		reqBody service.MintRequestBody
		resBody string
		code    int
	}{
		{
			desc: "valid request",
			reqBody: service.MintRequestBody{
				Conversion: service.Conversion{
					Beneficiary:  "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx",
					LiquidTxHash: "b356413f906468a3220f403c350d01a5880dbd1417f3ff294a4a2ff62faf0839",
					Descriptor:   "wpkh([6a00c946/0'/0'/501']02e24c96e967524fb2ad3b3e3c29c275e05934b12f420b7871443143d05ffe11c8)#8ktzldqn",
				},
				Signature: "ICucxAHOsf1kanl9UAjxMXemLmnP0deHWwyqdav68e8XCknJeaNBPFl9t7h52Ny1/XNgiQFu8XzrGLM8qahSy38=",
			},
			resBody: "",
			code:    200,
		},
		{
			desc:    "missing request fields",
			reqBody: service.MintRequestBody{},
			resBody: "{\"error\":\"Key: 'MintRequestBody.Conversion.Beneficiary' Error:Field validation for 'Beneficiary' failed on the 'required' tag\\nKey: 'MintRequestBody.Conversion.LiquidTxHash' Error:Field validation for 'LiquidTxHash' failed on the 'required' tag\\nKey: 'MintRequestBody.Conversion.Descriptor' Error:Field validation for 'Descriptor' failed on the 'required' tag\\nKey: 'MintRequestBody.Signature' Error:Field validation for 'Signature' failed on the 'required' tag\"}",
			code:    400,
		},
		{
			desc: "invalid signature",
			reqBody: service.MintRequestBody{
				Conversion: service.Conversion{
					Beneficiary:  "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx",
					LiquidTxHash: "b356413f906468a3220f403c350d01a5880dbd1417f3ff294a4a2ff62faf0123",
					Descriptor:   "wpkh([6a00c946/0'/0'/501']02e24c96e967524fb2ad3b3e3c29c275e05934b12f420b7871443143d05ffe11c8)#8ktzldqn",
				},
				Signature: "IKRJ47JiwZ/XR9anRworW2VoUbOWo+MM/7MO9fccp1u5fTv9gSsk5iQmgM3WEgvv2SeiOxdGDao4FONyXG5Xe3s=",
			},
			resBody: "{\"error\":\"invalid signature\"}",
			code:    400,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			bodyBytes, err := json.Marshal(tc.reqBody)
			assert.NoError(t, err)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/mint", bytes.NewBuffer(bodyBytes))
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
			assert.Equal(t, tc.resBody, w.Body.String())
		})
	}
}

func TestGetReceiveAddressRoute(t *testing.T) {
	t.Parallel()
	//_, router, pmClientMock, eClientMock, _ := setupService(t)
	router := gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock := testutil.NewMockIElementsClient(ctrl)

	db, err := leveldb.OpenFile("./conversions.db", nil)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	_ = service.NewR2PService(router, pmClientMock, eClientMock, db)
	// return

	pmClientMock.EXPECT().IsLegitMachine(gomock.Any()).Return(&testutil.IsLegitMachine, nil).AnyTimes()

	eClientMock.EXPECT().GetTransaction(gomock.Any(), gomock.Any()).Return(testutil.GetTransactionResult, nil).AnyTimes()
	eClientMock.EXPECT().DeriveAddresses(gomock.Any(), gomock.Any()).Return(testutil.DeriveAddressesResult, nil).AnyTimes()
	eClientMock.EXPECT().GetNewAddress(gomock.Any(), gomock.Any()).Return(testutil.ConfidentialAddr, nil).AnyTimes()
	eClientMock.EXPECT().GetAddressInfo(gomock.Any(), gomock.Any()).Return(testutil.AddressInfo, nil).AnyTimes()

	tests := []struct {
		desc              string
		planetmintAddress string
		resBody           service.AddressBody
		code              int
		errorMsg          string
	}{
		{
			desc:              "valid request",
			planetmintAddress: testutil.PlanetmintAddress,
			resBody: service.AddressBody{
				LiquidAddress:         testutil.ConfidentialAddr,
				PlanetmintBeneficiary: testutil.PlanetmintAddress,
			},
			code:     200,
			errorMsg: "",
		},
		{
			desc:              "missing request fields",
			planetmintAddress: "",
			resBody:           service.AddressBody{},
			code:              404,
			errorMsg:          "404 page not found",
		},
		{
			desc:              "Invalid planetmint machine address",
			planetmintAddress: "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx",
			resBody:           service.AddressBody{},
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
				var result service.AddressBody
				err = json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tc.resBody, result)
			}
		})
	}
}
