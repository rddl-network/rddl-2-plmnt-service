package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/planetmint/planetmint-go/utils"
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
		[]string{strconv.Itoa(int(cfg.Confirmations)), "false", "true", conversion.ConfidentialAddress, cfg.AcceptedAsset})

	if len(txDetails) != 1 {
		//create error that there are too much txinformations
		msg := "error: the received account information contains too much or not enough information"
		fmt.Println(msg)
		err = errors.New(msg)
		return
	}
	if len(txDetails[0].TxIDs) != 1 {
		//create error that there are too much transactions
		msg := "error: the account received more than 1 transaction"
		fmt.Println(msg)
		err = errors.New(msg)
		return
	}
	fmt.Printf("Result: " + txDetails[0].TxIDs[0])
	liquidTxHash := txDetails[0].TxIDs[0]

	// check if mint request has already been issued
	code, err := r2p.checkMintRequest(liquidTxHash)
	if err != nil {
		msg := "error while checking mint request: " + err.Error() + " code: " + strconv.Itoa(code)
		fmt.Println(msg)
		err = errors.New(msg)
		if code == http.StatusConflict {
			deleteEntry = true
		}
		return
	}

	convertedAmount := utils.RDDLToken2Uint(txDetails[0].Amount)
	plmntAmount := r2p.getConversion(convertedAmount)
	err = r2p.pmClient.MintPLMNT(conversion.PlanetmintAddress, plmntAmount, liquidTxHash)
	if err != nil {
		msg := "error while minting token: " + err.Error()
		fmt.Println(msg)
		err = errors.New(msg)
	}

	return
}

func (r2p *R2PService) checkMintRequest(liquidTxHash string) (code int, err error) {
	// check whether mint request already exists
	mr, err := r2p.pmClient.CheckMintRequest(liquidTxHash)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error while fetching mint request: %w", err)
	}

	// return because mint request for txhash is already
	if mr != nil {
		return http.StatusConflict, errors.New("already minted")
	}
	return
}
