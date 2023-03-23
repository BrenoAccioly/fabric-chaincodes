package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

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

	return ctx.GetStub().PutState(id, assetJSON)
}

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

// TODO event if sensor value reaches some limit, or something like this

func main() {
	chaincode, err := contractapi.NewChaincode(&SensorContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
