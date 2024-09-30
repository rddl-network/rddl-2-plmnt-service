package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/planetmint/planetmint-go/util"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

func (r2p *R2PService) registerPeriodicTasks() {
	r2p.tickerList = append(r2p.tickerList, time.NewTicker(2*time.Hour))
	r2p.tickerList = append(r2p.tickerList, time.NewTicker(2*time.Minute))
	go func() {
		for {
			select {
			case <-r2p.tickerList[0].C:
				go r2p.cleanupDB()
			case <-r2p.tickerList[1].C:
				go r2p.convertArrivedFunds()
			}
		}
	}()
}

func (r2p *R2PService) ExecutePotentialConversion(conversion ConversionRequest) (deleteEntry bool, err error) {
	cfg := config.GetConfig()
	txDetails, err := r2p.eClient.ListReceivedByAddress(cfg.GetElementsURL(),
		[]string{strconv.Itoa(int(cfg.Confirmations)), "false", "true", `"` + conversion.ConfidentialAddress + `"`, `"` + cfg.AcceptedAsset + `"`})
	if err != nil {
		msg := "error: invalid call to rpc with address " + conversion.ConfidentialAddress + " : " + err.Error()
		r2p.logger.Error("error", msg)
		err = errors.New(msg)
		return
	}
	if len(txDetails) == 0 {
		msg := "the address hasn't received any transactions for the given asset: %s - %s"
		r2p.logger.Debug("msg", fmt.Sprintf(msg, conversion.ConfidentialAddress, cfg.AcceptedAsset))
		return
	} else if len(txDetails) > 1 {
		msg := "the tx details for the address are unexpected: " + conversion.ConfidentialAddress
		r2p.logger.Error("error", msg)
		err = errors.New(msg)
		return
	}
	if len(txDetails[0].TxIDs) > 1 {
		// create error that there are too much transactions
		msg := "error: the account received more than 1 transaction: " + conversion.ConfidentialAddress
		r2p.logger.Error("error", msg)
		err = errors.New(msg)
		return
	}
	r2p.logger.Info("msg", "Conversion: "+conversion.ConfidentialAddress+" received tx: "+txDetails[0].TxIDs[0])
	liquidTxHash := txDetails[0].TxIDs[0]

	// check if mint request has already been issued
	code, err := r2p.checkMintRequest(liquidTxHash)
	if err != nil {
		msg := "error while checking mint request: " + err.Error() + " code: " + strconv.Itoa(code) + " for address " + conversion.ConfidentialAddress
		r2p.logger.Error("error", msg)
		err = errors.New(msg)
		return
	} else if code == http.StatusConflict {
		deleteEntry = true
		msg := "tx " + liquidTxHash + " got already minted"
		r2p.logger.Debug("msg", msg)
		return
	}

	convertedAmount := util.RDDLToken2Uint(txDetails[0].Amount)
	plmntAmount := GetConversion(convertedAmount)
	err = r2p.pmClient.MintPLMNT(conversion.PlanetmintAddress, plmntAmount, liquidTxHash)
	if err != nil {
		msg := "error while minting " + strconv.FormatUint(plmntAmount, 10) + " tokens (tx id " + liquidTxHash + ") for address " + conversion.PlanetmintAddress
		r2p.logger.Error("msg", msg)
		err = errors.New(msg)
	}

	return
}

func (r2p *R2PService) checkMintRequest(liquidTxHash string) (code int, err error) {
	// check whether mint request already exists
	mr, err := r2p.pmClient.CheckMintRequest(liquidTxHash)
	if err != nil {
		r2p.logger.Error("msg", "error while fetching mint request: "+err.Error())
		code = http.StatusInternalServerError
		err = fmt.Errorf("error while fetching mint request: %w", err)
		return
	}

	// return because mint request for txhash is already
	if mr != nil {
		r2p.logger.Debug("msg", "mint request: txid "+liquidTxHash+" got minted before ("+mr.String()+")")
		code = http.StatusConflict
		return
	}
	return
}
