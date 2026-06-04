/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	// "crypto/sha256"
	// "encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
    "github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
    "github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}
// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetId string, subId string, assetHash string) (map[string]interface{}, error) {
    fmt.Println("--- Start Create Asset ---")

    // 3. Get transaction ID
    txID := ctx.GetStub().GetTxID()

    // 4. Prepare the full asset record (stored in ledger)
    record := map[string]interface{}{
        "assetId":       assetId,
        "sub_id":        subId,
        "assetHash":     assetHash,
        "transactionID": txID,
    }

    assetJSON, err := json.Marshal(record)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal asset: %v", err)
    }

    // 5. Store using assetId as CouchDB document _id (via PutState key)
    if err := ctx.GetStub().PutState(assetId, assetJSON); err != nil {
        return nil, fmt.Errorf("failed to store asset: %v", err)
    }

    fmt.Printf("Stored asset with _id (assetId): %s\n", assetId)
    fmt.Println("--- Asset Created Successfully ---")

    // 6. Return only assetHash and txID
    return map[string]interface{}{
        "assetId":       assetId,
        "assetHash":     assetHash,
        "transactionID": txID,
    }, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetId string) (map[string]interface{}, error) {
    fmt.Println("--- Start Query Asset ---")

    assetJSON, err := ctx.GetStub().GetState(assetId)
    if err != nil {
        return nil, fmt.Errorf("Failed to read from world state: %v", err)
    }

    if assetJSON == nil {
        errMsg := fmt.Sprintf("The asset with assetId %s does not exist", assetId)
        fmt.Println(errMsg)
        fmt.Println("--- Queried Asset Unsuccessful ---\n")
        return nil, fmt.Errorf(errMsg)
    }

    var asset map[string]interface{}
    if err := json.Unmarshal(assetJSON, &asset); err != nil {
        return nil, fmt.Errorf("Failed to unmarshal asset data: %v", err)
    }

    // Format asset for logging
    formattedAsset, err := json.MarshalIndent(asset, "", "  ")
    if err != nil {
        fmt.Printf("Asset exists but failed to format for logging: %v\n", err)
    } else {
        fmt.Printf("Asset details:\n%s\n", formattedAsset)
    }

    fmt.Println("--- Queried Asset Successfully ---\n")
    return asset, nil
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, assetId string, newAssetHash string) (string, error) {
    fmt.Println("----- Start Update Asset -----")

    // Get existing asset directly by assetID key
    existingAssetJSON, err := ctx.GetStub().GetState(assetId)
    if err != nil {
        return "", fmt.Errorf("failed to read asset: %v", err)
    }
    if existingAssetJSON == nil {
        return "", fmt.Errorf("asset with assetId %s not found", assetId)
    }

    // Unmarshal existing asset
    var existingAsset map[string]interface{}
    if err := json.Unmarshal(existingAssetJSON, &existingAsset); err != nil {
        return "", fmt.Errorf("failed to unmarshal existing asset: %v", err)
    }

    // Preserve the original transactionID
    var prevTxID string
    if v, ok := existingAsset["transactionID"].(string); ok {
        prevTxID = v
        existingAsset["previousTransactionID"] = v
    } else {
        return "", fmt.Errorf("transactionID is missing or not a string in existing asset")
    }

    // Update asset fields
    existingAsset["assetHash"] = newAssetHash
    existingAsset["transactionID"] = ctx.GetStub().GetTxID()

    // Marshal and save updated asset
    updatedJSON, err := json.Marshal(existingAsset)
    if err != nil {
        return "", fmt.Errorf("failed to marshal updated asset: %v", err)
    }

    if err := ctx.GetStub().PutState(assetId, updatedJSON); err != nil {
        return "", fmt.Errorf("failed to update asset in ledger: %v", err)
    }

    // Create response object and marshal to JSON
    response := map[string]interface{}{
        "assetHash":             existingAsset["assetHash"],
        "transactionID":         existingAsset["transactionID"],
        "assetId":               assetId,
        "previousTransactionID": prevTxID,
    }

    responseJSON, err := json.Marshal(response)
    if err != nil {
        return "", fmt.Errorf("failed to marshal response: %v", err)
    }

    fmt.Println("----- Updated Asset Successfully -----")
    return string(responseJSON), nil
}

// DeleteAsset deletes an asset from the world state by assetId.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, assetId string) error {
    fmt.Printf("--- Start DeleteAsset for assetId: %s ---\n", assetId)

    // Check if the asset exists
    exists, err := s.AssetExists(ctx, assetId)
    if err != nil {
        return fmt.Errorf("Error checking asset existence: %v", err)
    }
    if !exists {
        return fmt.Errorf("Asset %s does not exist", assetId)
    }

    // Delete the asset
    err = ctx.GetStub().DelState(assetId)
    if err != nil {
        return fmt.Errorf("Failed to delete asset: %v", err)
    }

    fmt.Printf("--- Deleted assetId: %s successfully ---\n", assetId)
    return nil
}

// AssetExists checks if asset exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, assetId string) (bool, error) {
    data, err := ctx.GetStub().GetState(assetId)
    if err != nil {
        return false, err
    }
    return data != nil, nil
}

// ListAssetsBySubID returns all assets that match the given sub_id
func (s *SmartContract) ListAssetsBySubID(ctx contractapi.TransactionContextInterface, sub_id string) ([]map[string]interface{}, error) {
    fmt.Println("--- Start List Assets by Sub ID ---")
    fmt.Printf("Searching for assets with sub_id: %s\n", sub_id)

    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        fmt.Printf("Failed to get state iterator: %v\n", err)
        return nil, fmt.Errorf("Failed to retrieve assets: %v", err)
    }
    defer resultsIterator.Close()

    assets := make([]map[string]interface{}, 0)

    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            fmt.Printf("Error iterating through results: %v\n", err)
            return nil, fmt.Errorf("Failed to iterate through assets: %v", err)
        }

        if len(queryResponse.Key) > 0 && queryResponse.Key[0] == 0x00 {
            continue
        }

        var asset map[string]interface{}
        if err := json.Unmarshal(queryResponse.Value, &asset); err != nil {
            fmt.Printf("Failed to unmarshal asset with key %s: %v\n", queryResponse.Key, err)
            continue
        }

        if assetSubID, ok := asset["sub_id"].(string); ok && assetSubID == sub_id {
            assets = append(assets, asset)
        }
    }

    if len(assets) == 0 {
        fmt.Printf("No assets found with sub_id: %s\n", sub_id)
    } else {
        fmt.Printf("Found %d assets with sub_id: %s\n", len(assets), sub_id)
        for i, asset := range assets {
            tidyAsset, err := json.MarshalIndent(asset, "", "  ")
            if err != nil {
                fmt.Printf("Asset %d: Failed to format for logging\n", i+1)
            } else {
                fmt.Printf("Asset %d:\n%s\n", i+1, string(tidyAsset))
            }
        }
    }

    fmt.Println("--- List Assets by Sub ID Successfully ---\n")
    return assets, nil
}



func main() {
    cc, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        log.Panicf("Error creating chaincode: %v", err)
    }

    server := &shim.ChaincodeServer{
        CCID:    os.Getenv("CHAINCODE_ID"),
        Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
        CC:      cc,
        TLSProps: shim.TLSProperties{Disabled: true},
    }

    if err := server.Start(); err != nil {
        log.Panicf("Error starting chaincode server: %v", err)
    }
}

func (s *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetId string) ([]map[string]interface{}, error) {
    fmt.Println("--- Start Get Asset History ---")

    iterator, err := ctx.GetStub().GetHistoryForKey(assetId)
    if err != nil {
        return nil, fmt.Errorf("failed to get history: %v", err)
    }
    defer iterator.Close()

    var history []map[string]interface{}
    for iterator.HasNext() {
        record, err := iterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate history: %v", err)
        }

        entry := map[string]interface{}{
            "txId":      record.TxId,
            "timestamp": record.Timestamp.String(),
            "isDelete":  record.IsDelete,
        }

        if !record.IsDelete {
            var val map[string]interface{}
            if err := json.Unmarshal(record.Value, &val); err == nil {
                entry["value"] = val
            }
        }

        history = append(history, entry)
    }

    fmt.Println("--- Get Asset History Completed ---")
    return history, nil
}

