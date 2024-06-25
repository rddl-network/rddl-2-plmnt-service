package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type ConversionRequest struct {
	ConfidentialAddress string `binding:"required" json:"confidential-address"`
	PlanetmintAddress   string `binding:"required" json:"planetmint-address"`
	Timestamp           int64  `binding:"required" json:"timestamp"`
}

func (r2p *R2PService) addConversionRequest(confidentialAddress string, planetmintAddress string) (err error) {
	// store receive address - planetmint address pair
	var convReq ConversionRequest
	convReq.ConfidentialAddress = confidentialAddress
	convReq.PlanetmintAddress = planetmintAddress
	now := time.Now()
	convReq.Timestamp = now.Unix()

	convReqBytes, err := json.Marshal(convReq)
	if err != nil {
		r2p.logger.Error("error", "Error serializing ConversionRequest: "+err.Error())
		return
	}

	r2p.dbMutex.Lock()
	err = r2p.db.Put([]byte(confidentialAddress), convReqBytes, nil)
	r2p.dbMutex.Unlock()
	if err != nil {
		r2p.logger.Error("error", "storing addresses in DB: "+err.Error())
		return
	}
	return
}

func (r2p *R2PService) deleteEntry(key []byte) (err error) {
	r2p.dbMutex.Lock()
	err = r2p.db.Delete(key, nil)
	r2p.dbMutex.Unlock()
	return
}

func (r2p *R2PService) cleanupDB() {
	// Create an iterator for the database
	iter := r2p.db.NewIterator(nil, nil)
	defer iter.Release() // Make sure to release the iterator at the end

	// Iterate over all elements in the database
	for iter.Next() {
		// Use iter.Key() and iter.Value() to access the key and value
		key := iter.Key()
		value := iter.Value()
		var req ConversionRequest
		err := json.Unmarshal(value, &req)
		if err != nil {
			log.Printf("Failed to unmarshal entry: %s - %v", string(key), err)
			continue
		}
		now := time.Now()
		if now.Unix()-req.Timestamp > int64((12 * time.Hour).Seconds()) {
			// If the entry is older than 12 hours, delete it
			err := r2p.deleteEntry(key)
			if err != nil {
				log.Printf("Failed to delete entry: %v", err)
			}
		}
	}

	// Check for any errors encountered during iteration
	if err := iter.Error(); err != nil {
		r2p.logger.Error("error", err.Error())
	}
}

func (r2p *R2PService) convertArrivedFunds() {
	// Create an iterator for the database, nil means the whole database
	iter := r2p.db.NewIterator(nil, nil)
	defer iter.Release()

	// Start from the last key
	for iter.Last(); iter.Valid(); iter.Prev() {
		key := iter.Key()
		value := iter.Value()
		msg := fmt.Sprintf("Key: %s, Value: %s\n", key, value)
		r2p.logger.Info("msg", msg)
		var req ConversionRequest
		err := json.Unmarshal(value, &req)
		if err != nil {
			r2p.logger.Error("error", fmt.Sprintf("Failed to unmarshal entry: %s - %v", string(key), err))
			continue
		}
		deleteEntry, err := r2p.ExecutePotentialConversion(req)
		if err != nil {
			r2p.logger.Error("error", fmt.Sprintf("Failed to convert entry: %s - %v", string(key), err))
			if deleteEntry {
				r2p.logger.Info("msg", fmt.Sprintf("delete entry: %s ", string(key)))
				err = r2p.deleteEntry(key)
				if err != nil {
					r2p.logger.Error("error", fmt.Sprintf("deletion of entry %s failed: %s", string(key), err.Error()))
				}
			}
		}
	}
	// Check for any errors found during iteration
	if err := iter.Error(); err != nil {
		log.Println(err.Error())
	}
}
