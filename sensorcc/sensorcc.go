package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const MAX_VALUE_LIMIT = 100

type SensorContract struct {
	contractapi.Contract
}

type SensorAsset struct {
	ID        string `json:id`
	Timestamp string `json:timestamp`
	Value     int    `json:value`
}

func (s *SensorContract) HasAsset(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	sensorAsset, err := ctx.GetStub().GetState(id)

	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	return sensorAsset != nil, nil
}

/*
Creates or updates a sensor asset. Emits an event if the value reaches a limit
*/
func (s *SensorContract) PutAsset(
	ctx contractapi.TransactionContextInterface,
	id string,
	ts string,
	value int,
) error {
	sensorAsset := SensorAsset{
		ID:        id,
		Timestamp: ts,
		Value:     value,
	}

	assetJSON, err := json.Marshal(sensorAsset)

	if err != nil {
		return err
	}

	if value > MAX_VALUE_LIMIT {
		ctx.GetStub().SetEvent("max-value-limit", []byte(fmt.Sprintf("Max limit exceeded in sensor %s, value %d", id, value)))
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

/*
Deletes a sensor asset
*/
func (s *SensorContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exist, err := s.HasAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	if !exist {
		return fmt.Errorf("Asset with id: %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

/*
Gets a sensor asset given its id
*/
func (s *SensorContract) GetAsset(ctx contractapi.TransactionContextInterface, id string) (*SensorAsset, error) {
	sensorAssetJSON, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	if sensorAssetJSON == nil {
		return nil, fmt.Errorf("Asset with id: %s does not exist", id)
	}

	var sensorAsset SensorAsset
	err = json.Unmarshal(sensorAssetJSON, &sensorAsset)
	if err != nil {
		return nil, err
	}

	return &sensorAsset, nil
}

/*
Gets multiple assets given a JSON array of ids (Not Working)
*/
func (s *SensorContract) GetAssets(ctx contractapi.TransactionContextInterface, idsJSON string) (assets []*SensorAsset, err error) {

	if !json.Valid([]byte(idsJSON)) {
		return nil, fmt.Errorf("Invalid JSON object as argument")
	}

	var ids []string
	err = json.Unmarshal([]byte(idsJSON), &ids)

	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		assetJSON, err := ctx.GetStub().GetState(id)

		if err != nil {
			return nil, err
		}

		var sensorAsset SensorAsset
		err = json.Unmarshal(assetJSON, &sensorAsset)

		if err != nil {
			return nil, err
		}

		assets = append(assets, &sensorAsset)
	}

	return
}

/*
Gets all sensor assets
*/
func (s *SensorContract) GetAllAssets(ctx contractapi.TransactionContextInterface) (assets []*SensorAsset, err error) {
	stateIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}

	defer stateIterator.Close()

	for stateIterator.HasNext() {
		assetJSON, err := stateIterator.Next()

		if err != nil {
			return nil, err
		}

		var sensorAsset SensorAsset
		err = json.Unmarshal(assetJSON.Value, &sensorAsset)

		if err != nil {
			return nil, err
		}

		assets = append(assets, &sensorAsset)
	}

	return
}

/*
Gets a sensor asset state history given its 'id'. Returns only the 'max' most recent states
*/
func (s *SensorContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, id string, max int) (assets []*SensorAsset, err error) {
	historyIterator, err := ctx.GetStub().GetHistoryForKey(id)

	if err != nil {
		return nil, err
	}

	defer historyIterator.Close()

	for historyIterator.HasNext() && len(assets) < max {
		assetJSON, err := historyIterator.Next()

		if err != nil {
			return nil, err
		}

		var sensorAsset SensorAsset
		err = json.Unmarshal(assetJSON.Value, &sensorAsset)

		if err != nil {
			return nil, err
		}

		assets = append(assets, &sensorAsset)

	}

	return
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SensorContract{})
	if err != nil {
		log.Panicf("Error creating sensor chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting sensor chaincode: %v", err)
	}
}
