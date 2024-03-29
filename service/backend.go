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
		fmt.Printf("Error serializing ConversionRequest: %v", err)
		return
	}

	r2p.dbMutex.Lock()
	err = r2p.db.Put([]byte(confidentialAddress), convReqBytes, nil)
	r2p.dbMutex.Unlock()
	if err != nil {
		fmt.Println("storing addresses in DB: " + err.Error())
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
		fmt.Println(err.Error())
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

		fmt.Printf("Key: %s, Value: %s\n", key, value)
		var req ConversionRequest
		err := json.Unmarshal(value, &req)
		if err != nil {
			log.Printf("Failed to unmarshal entry: %s - %v", string(key), err)
			continue
		}
		deleteEntry, err := r2p.ExecutePotentialConversion(req)
		if err != nil {
			log.Printf("Failed to convert entry: %s - %v", string(key), err)
			if deleteEntry {
				log.Printf("delete entry: %s ", string(key))
				err = r2p.deleteEntry(key)
				if err != nil {
					log.Printf("deletion of entry %s failed: %s", string(key), err.Error())
				}
			}
		}
	}
	// Check for any errors found during iteration
	if err := iter.Error(); err != nil {
		log.Println(err.Error())
	}
}
