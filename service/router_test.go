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

func setupService(t *testing.T) (app *service.R2PService, router *gin.Engine) {
	router = gin.Default()
	ctrl := gomock.NewController(t)
	pmClientMock := testutil.NewMockIPlanetmintClient(ctrl)
	elements.Client = &elementsmocks.MockClient{} // add mock request for get tx for testing
	app = service.NewR2PService(router, pmClientMock)
	return
}

func TestPostMintRequestRoute(t *testing.T) {
	t.Parallel()
	_, router := setupService(t)

	tests := []struct {
		desc    string
		reqBody service.MintRequestBody
		resBody string
		code    int
	}{}

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
