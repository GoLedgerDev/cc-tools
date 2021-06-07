package transactions

import (
	"encoding/json"
	"time"

	"github.com/goledgerdev/cc-tools/assets"
	"github.com/goledgerdev/cc-tools/errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ReadAssetHistory fetches an asset key history from the blockchain
var ReadAssetHistory = Transaction{
	Tag:         "readAssetHistory",
	Label:       "Read Asset History",
	Description: "",
	Method:      "GET",

	MetaTx: true,
	Args: []Argument{
		{
			Tag:         "key",
			Description: "Key of the asset to be read.",
			DataType:    "@key",
			Required:    true,
		},
		{
			Tag:         "timeTarget",
			Description: "Optional parameter to retrieve specific version of the asset.",
			DataType:    "datetime",
		},
	},
	ReadOnly: true,
	Routine: func(stub shim.ChaincodeStubInterface, req map[string]interface{}) ([]byte, errors.ICCError) {
		// This is safe to do because validation is done before calling routine
		key := req["key"].(assets.Key)
		timeTarget := req["timeTarget"]

		// Get asset's history from blockchain
		historyIterator, err := stub.GetHistoryForKey(key.Key())
		if err != nil {
			return nil, errors.WrapError(err, "failed to read asset from blockchain")
		}
		if historyIterator == nil {
			return nil, errors.NewCCError("history not found", 404)
		}
		defer historyIterator.Close()

		if !historyIterator.HasNext() {
			return nil, errors.NewCCError("history not found", 404)
		}

		if timeTarget == nil {
			response := make([]map[string]interface{}, 0)
			for historyIterator.HasNext() {
				queryResponse, err := historyIterator.Next()
				if err != nil {
					return nil, errors.WrapError(err, "error iterating response")
				}

				data := make(map[string]interface{})

				if queryResponse.IsDelete {
					data["_isDelete"] = queryResponse.IsDelete
				} else {
					err = json.Unmarshal(queryResponse.Value, &data)
					if err != nil {
						return nil, errors.WrapError(err, "failed to unmarshal queryResponse's values")
					}
				}
				data["_timestamp"] = time.Unix(queryResponse.Timestamp.Seconds, int64(queryResponse.Timestamp.Nanos)).Format(time.RFC3339)
				response = append(response, data)
			}
			responseJSON, err := json.Marshal(response)
			if err != nil {
				return nil, errors.WrapError(err, "error marshaling response")
			}

			return responseJSON, nil
		} else {
			response := make(map[string]interface{})
			target := timeTarget.(time.Time)
			closestTime := time.Time{}

			for historyIterator.HasNext() {
				queryResponse, err := historyIterator.Next()
				if err != nil {
					return nil, errors.WrapError(err, "error iterating response")
				}

				timestamp := queryResponse.Timestamp.AsTime()
				if timestamp.Before(target) && timestamp.After(closestTime) {
					closestTime = timestamp
					if !queryResponse.IsDelete {
						err = json.Unmarshal(queryResponse.Value, &response)
						if err != nil {
							return nil, errors.WrapError(err, "failed to unmarshal queryResponse's values")
						}
					}
					response["_isDelete"] = queryResponse.IsDelete
					response["_timestamp"] = timestamp.Format(time.RFC3339)
				}
			}

			responseJSON, err := json.Marshal(response)
			if err != nil {
				return nil, errors.WrapError(err, "error marshaling response")
			}

			return responseJSON, nil
		}
	},
}
