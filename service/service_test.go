package service_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
	"github.com/stretchr/testify/assert"
)

func TestValidateConversionSignature(t *testing.T) {
	app, _, _, _ := setupService(t)

	sk := secp256k1.GenPrivKey()

	cv := service.Conversion{
		Beneficiary:  "beneficiary",
		LiquidTXHash: "liquidtxhash",
	}

	cvBytes, _ := json.Marshal(cv)

	signature, err := sk.Sign(cvBytes)
	assert.NoError(t, err)
	signatureHex := hex.EncodeToString(signature)

	pk := sk.PubKey()
	pkHex := hex.EncodeToString(pk.Bytes())

	isValid, err := app.ValidateConversionSignature(cv, signatureHex, pkHex)
	assert.NoError(t, err)

	assert.True(t, isValid)
}
