package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rddl-network/rddl-2-plmnt-service/client"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/stretchr/testify/assert"
)

func TestGetReceiveAddress(t *testing.T) {
	t.Parallel()

	expectedRes := service.ReceiveAddressResponse{
		LiquidAddress:         "liquidAddress",
		PlanetmintBeneficiary: "plmntAddress",
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/receiveaddress/"+expectedRes.PlanetmintBeneficiary, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		bytes, err := json.Marshal(expectedRes)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(bytes)
		assert.NoError(t, err)
	}))
	defer mockServer.Close()

	c := client.NewR2PClient(mockServer.URL, mockServer.Client())
	res, err := c.GetReceiveAddress(context.Background(), expectedRes.PlanetmintBeneficiary)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes.PlanetmintBeneficiary, res.PlanetmintBeneficiary)
	assert.Equal(t, expectedRes.LiquidAddress, res.LiquidAddress)
}
