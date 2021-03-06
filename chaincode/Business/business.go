package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type chainCode struct {
}

type businessInfo struct {
	BusinessName                         string `json:"Name"`
	BusinessAcNo                         string `json:"AcNo"`
	BusinessLimit                        int64  `json:"Limit"`
	BusinessWalletID                     string `json:"MainWallet"`      //will take the values for the respective wallet from the user
	BusinessLoanWalletID                 string `json:"LoanWallet"`      //will take the values for the respective wallet from the user
	BusinessLiabilityWalletID            string `json:"LiabilityWallet"` //will take the values for the respective wallet from the user
	MaxROI                               int64  `json:"MaxROI"`
	MinROI                               int64  `json:"MinROI"`
	BusinessPrincipalOutstandingWalletID string `json:"POsWallet"` //will take the values for the respective wallet from the user
	BusinessChargesOutstandingWalletID   string `json:"COsWallet"` //will take the values for the respective wallet from the user
}

// toChaincodeArgs returns byte arrau of string of arguments, so it can be passed to other chaincodes
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func (c *chainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	bis := businessInfo{}
	indexName := "BusinessAcNo~BusinessName"
	acntNoNameKey, err := stub.CreateCompositeKey(indexName, []string{bis.BusinessAcNo, bis.BusinessName})
	if err != nil {
		return shim.Error("businesscc: " + "Unable to create composite key BusinessAcNo~BusinessName in businesscc")
	}
	value := []byte{0x00}
	stub.PutState(acntNoNameKey, value)
	return shim.Success(nil)
}

func (c *chainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "putNewBusinessInfo" {
		//Creates a new Business Information
		return putNewBusinessInfo(stub, args)
	} else if function == "getBusinessInfo" {
		//Retrieves the Business information
		return getBusinessInfo(stub, args)
	} else if function == "getWalletID" {
		//Returns the walletID for the required wallet type
		return getWalletID(stub, args)
	} else if function == "bisIDexists" {
		//To check the BusinessId existence
		return bisIDexists(stub, args[0])
	} else if function == "updateBusinessInfo" {
		//Updates Business Limit / MAX ROI / MAX ROI if required
		return updateBusinessInfo(stub, args)
	}
	return shim.Error("businesscc: " + "No function named " + function + " in Businessssssss")
}

func putNewBusinessInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 11 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("businesscc: " + "Invalid number of arguments in putNewBusinessInfo (required:11) given:" + xLenStr)

	}

	response := bisIDexists(stub, args[0])
	if response.Status != shim.OK {
		return shim.Error("businesscc: " + response.Message)
	}

	businessLimitConv, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error("businesscc: " + err.Error())
	}
	if businessLimitConv <= 0 {
		return shim.Error("businesscc: " + "Invalid Business Limit value: " + args[3])
	}

	hash := sha256.New()

	// Hashing BusinessWalletID
	BusinessWalletStr := args[2] + "BusinessWallet"
	hash.Write([]byte(BusinessWalletStr))
	md := hash.Sum(nil)
	BusinessWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, BusinessWalletIDsha, args[4])

	// Hashing BusinessLoanWalletID
	BusinessLoanWalletStr := args[2] + "BusinessLoanWallet"
	hash.Write([]byte(BusinessLoanWalletStr))
	md = hash.Sum(nil)
	BusinessLoanWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, BusinessLoanWalletIDsha, args[5])

	// Hashing BusinessLiabilityWalletID
	BusinessLiabilityWalletStr := args[2] + "BusinessLiabilityWallet"
	hash.Write([]byte(BusinessLiabilityWalletStr))
	md = hash.Sum(nil)
	BusinessLiabilityWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, BusinessLiabilityWalletIDsha, args[6])

	maxROIconvertion, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		fmt.Printf("Invalid Maximum ROI: %s\n", args[7])
		return shim.Error("businesscc: " + err.Error())
	}
	if maxROIconvertion <= 0 {
		return shim.Error("businesscc: " + "Invalid Max ROI value: " + args[7])
	}

	minROIconvertion, err := strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		fmt.Printf("Invalid Minimum ROI: %s\n", args[8])
		return shim.Error("businesscc: " + err.Error())
	}
	if minROIconvertion <= 0 {
		return shim.Error("businesscc: " + "Invalid Min ROI value: " + args[8])
	}

	// Hashing BusinessPrincipalOutstandingWalletID
	BusinessPrincipalOutstandingWalletStr := args[2] + "BusinessPrincipalOutstandingWallet"
	hash.Write([]byte(BusinessPrincipalOutstandingWalletStr))
	md = hash.Sum(nil)
	BusinessPrincipalOutstandingWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, BusinessPrincipalOutstandingWalletIDsha, args[9])

	// Hashing BusinessChargesOutstandingWalletID
	BusinessInterestOutstandingWalletStr := args[2] + "BusinessInterestOutstandingWallet"
	hash.Write([]byte(BusinessInterestOutstandingWalletStr))
	md = hash.Sum(nil)
	BusinessChargesOutstandingWalletIDsha := hex.EncodeToString(md)
	createWallet(stub, BusinessChargesOutstandingWalletIDsha, args[10])

	newInfo := &businessInfo{args[1], args[2], businessLimitConv, BusinessWalletIDsha, BusinessLoanWalletIDsha, BusinessLiabilityWalletIDsha, maxROIconvertion, minROIconvertion, BusinessPrincipalOutstandingWalletIDsha, BusinessChargesOutstandingWalletIDsha}
	newInfoBytes, _ := json.Marshal(newInfo)
	err = stub.PutState(args[0], newInfoBytes) // businessID = args[0]
	if err != nil {
		return shim.Error("businesscc: " + err.Error())
	}

	fmt.Println("Successfully added buissness " + args[1] + " to the ledger")
	return shim.Success([]byte("Successfully added buissness " + args[1] + " to the ledger"))
}

func createWallet(stub shim.ChaincodeStubInterface, walletID string, amt string) pb.Response {
	//Calling the wallet Chaincode to create new wallet
	chaincodeArgs := toChaincodeArgs("newWallet", walletID, amt)
	response := stub.InvokeChaincode("walletcc", chaincodeArgs, "myc")
	if response.Status != shim.OK {
		return shim.Error("businesscc: " + "Unable to create new wallet from business")
	}
	return shim.Success([]byte("businesscc: " + "created new wallet from business"))
}

func getBusinessInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("businesscc: " + "Invalid number of arguments in getBusinessInfo (required:1) given:" + xLenStr)
	}

	parsedBusinessInfo := businessInfo{}
	businessIDvalue, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("businesscc: " + "Failed to get the business information: " + err.Error())
	} else if businessIDvalue == nil {
		return shim.Error("businesscc: " + "No information is avalilable on this businessID " + args[0])
	}

	err = json.Unmarshal(businessIDvalue, &parsedBusinessInfo)
	if err != nil {
		return shim.Error("businesscc: " + "Unable to parse businessInfo into the structure " + err.Error())
	}
	jsonString := fmt.Sprintf("%+v", parsedBusinessInfo)
	fmt.Printf("Business Info: %s\n", jsonString)
	return shim.Success(nil)
}

func bisIDexists(stub shim.ChaincodeStubInterface, bisID string) pb.Response {
	ifExists, _ := stub.GetState(bisID)
	if ifExists != nil {
		fmt.Println(ifExists)
		return shim.Error("businesscc: " + "BusinessId " + bisID + " exits. Cannot create new ID")
	}
	return shim.Success(nil)
}

func updateBusinessInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	/*
		args[0] -> BusinessId
		args[1] -> Business Limit / MAX ROI / MAX ROI
		args[2] -> values
	*/
	if len(args) != 3 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("businesscc: " + "Invalid number of arguments in updateBusinessInfo(business) (required:3) given:" + xLenStr)
	}

	parsedBusinessInfo := businessInfo{}
	businessIDvalue, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("businesscc: " + "Failed to get the business information(updateBusinessInfo): " + err.Error())
	} else if businessIDvalue == nil {
		return shim.Error("businesscc: " + "No information is avalilable on this (updateBusinessInfo) businessID " + args[0])
	}

	err = json.Unmarshal(businessIDvalue, &parsedBusinessInfo)
	if err != nil {
		return shim.Error("businesscc: " + "Unable to parse businessInfo into the structure(updateBusinessInfo) " + err.Error())
	}

	lowerStr := strings.ToLower(args[1])

	value, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("businesscc: " + "value (updateBusinessInfo):" + err.Error())
	}

	if lowerStr == "business limit" {
		parsedBusinessInfo.BusinessLimit = value
	} else if lowerStr == "max roi" {
		parsedBusinessInfo.MaxROI = value
	} else if lowerStr == "min roi" {
		parsedBusinessInfo.MinROI = value
	}

	parsedBusinessInfoBytes, _ := json.Marshal(parsedBusinessInfo)
	err = stub.PutState(args[0], parsedBusinessInfoBytes)
	if err != nil {
		return shim.Error("businesscc: " + "Error in updating business: " + err.Error())
	}

	return shim.Success([]byte("businesscc: " + "Successfully updated Business " + args[0]))

}

func getWalletID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		xLenStr := strconv.Itoa(len(args))
		return shim.Error("businesscc: " + "Invalid number of arguments in getWalletId(business) (required:2) given:" + xLenStr)
	}

	parsedBusinessInfo := businessInfo{}
	businessIDvalue, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("businesscc: " + "Failed to get the business information: " + err.Error())
	} else if businessIDvalue == nil {
		return shim.Error("businesscc: " + "No information is avalilable on this businessID " + args[0])
	}

	err = json.Unmarshal(businessIDvalue, &parsedBusinessInfo)
	if err != nil {
		return shim.Error("businesscc: " + "Unable to parse into the structure " + err.Error())
	}

	walletID := ""

	switch args[1] {
	case "main":
		walletID = parsedBusinessInfo.BusinessWalletID
	case "loan":
		walletID = parsedBusinessInfo.BusinessLoanWalletID
	case "liability":
		walletID = parsedBusinessInfo.BusinessLiabilityWalletID
	case "principalOut":
		walletID = parsedBusinessInfo.BusinessPrincipalOutstandingWalletID
	case "chargesOut":
		walletID = parsedBusinessInfo.BusinessChargesOutstandingWalletID
	default:
		return shim.Error("businesscc: " + "There is no wallet of this type in Business :" + args[1])
	}

	return shim.Success([]byte(walletID))
}

func main() {
	err := shim.Start(new(chainCode))
	if err != nil {
		fmt.Printf("businesscc: "+"Error starting Business chaincode: %s\n", err)
	}

}
