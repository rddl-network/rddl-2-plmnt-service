package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	elements "github.com/rddl-network/elements-rpc"
	elementsmocks "github.com/rddl-network/elements-rpc/utils/mocks"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/rddl-network/rddl-2-plmnt-service/testutil"
	"github.com/stretchr/testify/assert"
)

func setupService(t *testing.T) (app *service.R2PService, router *gin.Engine, pmClientMock *testutil.MockIPlanetmintClient, eClientMock *testutil.MockIElementsClient) {
	router = gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock = testutil.NewMockIPlanetmintClient(ctrl)
	eClientMock = testutil.NewMockIElementsClient(ctrl)
	elements.Client = &elementsmocks.MockClient{}
	app = service.NewR2PService(router, pmClientMock, eClientMock)
	return
}

func TestPostMintRequestRoute(t *testing.T) {
	t.Parallel()
	_, router, pmClientMock, eClientMock := setupService(t)

	pmClientMock.EXPECT().CheckMintRequest(gomock.Any()).Return(nil, nil).AnyTimes()
	pmClientMock.EXPECT().MintPLMNT(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	eClientMock.EXPECT().GetTransaction(gomock.Any(), gomock.Any()).Return(testutil.GetTransactionResult, nil).AnyTimes()

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
					Beneficiary:  "beneficiary",
					LiquidTXHash: "liquidtxhash",
				},
				AddressDescriptor: "addressDescriptor",
				Signature:         "asd",
			},
			resBody: "",
			code:    200,
		},
		{
			desc:    "bad request",
			reqBody: service.MintRequestBody{},
			resBody: "{\"error\":\"Key: 'MintRequestBody.Conversion.Beneficiary' Error:Field validation for 'Beneficiary' failed on the 'required' tag\\nKey: 'MintRequestBody.Conversion.LiquidTXHash' Error:Field validation for 'LiquidTXHash' failed on the 'required' tag\\nKey: 'MintRequestBody.AddressDescriptor' Error:Field validation for 'AddressDescriptor' failed on the 'required' tag\\nKey: 'MintRequestBody.Signature' Error:Field validation for 'Signature' failed on the 'required' tag\"}",
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
