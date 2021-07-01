package transactions

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func invokeAndVerify(stub *shim.MockStub, txName string, req, expectedRes interface{}, expectedStatus int32) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
		return err
	}

	res := stub.MockInvoke(txName, [][]byte{
		[]byte(txName),
		reqBytes,
	})

	if res.GetStatus() != expectedStatus {
		log.Println(res.GetMessage())
		return fmt.Errorf("expected %d got %d", expectedStatus, res.GetStatus())
	}

	var resData interface{}
	if expectedStatus == 200 {
		err = json.Unmarshal(res.GetPayload(), &resData)
	} else {
		resData = res.GetMessage()
	}
	if err != nil {
		log.Println(err)
		return err
	}
	if !reflect.DeepEqual(resData, expectedRes) {
		log.Println("these should be equal")
		log.Printf("%#v\n", resData)
		log.Printf("%#v\n", expectedRes)
		return fmt.Errorf("unexpected response")
	}

	return nil
}

func isEmpty(stub *shim.MockStub, key string) bool {
	stub.MockTransactionStart("ensureDeletion")
	defer stub.MockTransactionEnd("ensureDeletion")
	state := stub.State[key]
	if state != nil {
		return false
	}
	return true
}
