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

func setupService(t *testing.T) (app *service.R2PService, router *gin.Engine, pmClientMock *testutil.MockIPlanetmintClient) {
	router = gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock = testutil.NewMockIPlanetmintClient(ctrl)
	elements.Client = &elementsmocks.MockClient{} // add mock request for get tx for testing
	app = service.NewR2PService(router, pmClientMock)
	return
}

func TestPostMintRequestRoute(t *testing.T) {
	t.Parallel()
	_, router, pmClientMock := setupService(t)

	pmClientMock.EXPECT().CheckMintRequest(gomock.Any()).Return(nil, nil).AnyTimes()
	pmClientMock.EXPECT().MintPLMNT(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
				Signature: "asd",
				PublicKey: "pubkey",
			},
			resBody: "body",
			code:    200,
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
