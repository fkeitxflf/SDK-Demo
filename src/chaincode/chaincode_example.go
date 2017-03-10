/*
	author: Andy & James
	time:16/7/05
	MIT License
	Change history:
	2017/03/07 --James
	1)Delete some commented code.
	2017/03/06 --James
	1) Change the key in the company struct, as fabric do not support int key type map.
	2) Change the Company ID as string type, which get from the args.
	3) Add changeContractState func to invoke func
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var transactionNo int = 0

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type CenterBank struct {
	Name        string
	TotalNumber int
	RestNumber  int
}

type Company struct {
	Name        string
	Type        int //0: CoreCompany, 1: loanCompany
	ID          string
	TotalNumber int
	RestNumber  int
	BankTotal   map[string]int
	BankRest    map[string]int
}

type FinancingContract struct {
	ContractId       string
	State            int // 0:pending 1:approved 2:rejected
	CoreCompanyID    string
	FinanceCompanyID string
	BankID           string
	Amount           int
}

type Transaction struct {
	FromType int //CenterBank 0 Bank 1  Company 1
	FromCpID string
	FromBkID string
	ToType   int //Bank 1 Company 2
	ToCpID   string
	ToBkID   string
	Time     int64
	Number   int
	ID       int
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things.
// Create Platform Center Bank , and issue some number of coin.

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	var totalNumber int
	var centerBank CenterBank
	var cbBytes []byte
	totalNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	centerBank = CenterBank{Name: "centerBank", TotalNumber: totalNumber, RestNumber: totalNumber}
	err = writeCenterBank(stub, centerBank)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	cbBytes, err = json.Marshal(&centerBank)
	if err != nil {
		return nil, errors.New("Error retrieving cbBytes")
	}
	return cbBytes, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "createCompany" {
		return t.createCompany(stub, args)
	} else if function == "issueCoin" {
		return t.issueCoin(stub, args)
	} else if function == "issueCoinToCp" {
		return t.issueCoinToCp(stub, args)
	} else if function == "changeContractState" {
		return t.changeContractState(stub, args)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	} else if function == "createFinancingContract" {
		return t.createFinancingContract(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) createCompany(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	var company Company
	var cpBytes []byte
	var typeInt int

	banktotal := make(map[string]int)
	banktotal["centerBank"] = 0
	bankrest := make(map[string]int)
	bankrest["centerBank"] = 0

	typeInt, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for company type.")
	}

	company = Company{Name: args[0], Type: typeInt, TotalNumber: 0, RestNumber: 0, BankTotal: banktotal, BankRest: bankrest, ID: args[2]}

	err = writeCompany(stub, company)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	cpBytes, err = json.Marshal(&company)
	if err != nil {
		return nil, err
	}

	return cpBytes, nil
}

func (t *SimpleChaincode) createFinancingContract(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}
	var financing FinancingContract
	var financingBytes []byte
	var amountNumber int
	var ContractId string
	ContractId = args[0]
	amountNumber, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	financing = FinancingContract{ContractId: args[0], State: 0, CoreCompanyID: args[1], FinanceCompanyID: args[2], BankID: args[3], Amount: amountNumber}

	err = writeFinancingContract(stub, ContractId, financing)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	financingBytes, err = json.Marshal(&financing)
	if err != nil {
		return nil, errors.New("Error retrieving cbBytes")
	}

	fmt.Println(string(financingBytes))
	return financingBytes, nil
}

func (t *SimpleChaincode) issueCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	var centerBank CenterBank
	var tsBytes []byte

	issueNumber, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("want Integer number")
	}
	centerBank, _, err = getCenterBank(stub)
	if err != nil {
		return nil, errors.New("get errors")
	}

	centerBank.TotalNumber = centerBank.TotalNumber + issueNumber
	centerBank.RestNumber = centerBank.RestNumber + issueNumber

	err = writeCenterBank(stub, centerBank)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	transaction := Transaction{FromType: 0, FromCpID: "0", FromBkID: "0", ToType: 0, ToCpID: "0", ToBkID: "0", Time: time.Now().Unix(), Number: issueNumber, ID: transactionNo}
	err = writeTransaction(stub, transaction)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	tsBytes, err = json.Marshal(&transaction)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}

	transactionNo = transactionNo + 1
	return tsBytes, nil
}

func (t *SimpleChaincode) issueCoinToCp(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	var centerBank CenterBank
	var company Company
	var companyId string
	var bankId string
	var issueNumber int
	var tsBytes []byte
	var err error

	companyId = args[0]
	bankId = args[1]

	issueNumber, err = strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("want Integer number")
	}

	centerBank, _, err = getCenterBank(stub)
	if err != nil {
		return nil, errors.New("get errors")
	}
	if centerBank.RestNumber < issueNumber {
		return nil, errors.New("Not enough money centerBank.RestNumber is " + string(centerBank.RestNumber) + ",issueNumber is " + string(issueNumber))
	}

	company, _, err = getCompanyById(stub, companyId)
	if err != nil {
		return nil, errors.New("get errors")
	}
	company.RestNumber += issueNumber
	company.TotalNumber += issueNumber
	centerBank.RestNumber = centerBank.RestNumber - issueNumber

	if _, ok := company.BankTotal[bankId]; ok {
		company.BankTotal[bankId] += issueNumber
		company.BankRest[bankId] += issueNumber
	} else {
		company.BankTotal[bankId] = issueNumber
		company.BankRest[bankId] = issueNumber
	}

	err = writeCenterBank(stub, centerBank)
	if err != nil {
		return nil, errors.New("write errors" + err.Error())
	}

	err = writeCompany(stub, company)
	if err != nil {
		err = writeCenterBank(stub, centerBank)
		if err != nil {
			return nil, errors.New("roll down errors" + err.Error())
		}
		return nil, err
	}

	transaction := Transaction{FromType: 0, FromCpID: "0", FromBkID: "0", ToType: 1, ToCpID: companyId, ToBkID: bankId, Time: time.Now().Unix(), Number: issueNumber, ID: transactionNo}
	err = writeTransaction(stub, transaction)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	tsBytes, err = json.Marshal(&transaction)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}

	transactionNo = transactionNo + 1
	return tsBytes, nil
}

/*
 * Name:changeContractStat
 * arg[0]: contractid
 * arg[1]: state -> 0 pending, 1 approved, 2 rejected.
 */

func (t *SimpleChaincode) changeContractState(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	var financing FinancingContract
	var ContractId string
	var state int
	var err error
	var financingBytes []byte

	ContractId = args[0]
	state, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("want Integer number")
	} else if state < 0 || state > 3 {
		return nil, errors.New("changeContractState(): state is out of range")
	}

	financing, _, err = getFinancingContractById(stub, ContractId)

	if err != nil {
		return nil, errors.New("changeContractState(): get errors")
	}

	if financing.State == state {
		financingBytes, err = json.Marshal(&financing)
		if err != nil {
			return financingBytes, nil
		} else {
			return nil, errors.New("Change Byte error")
		}
	} else {
		financing.State = state
	}

	err = writeFinancingContract(stub, ContractId, financing)
	if err != nil {
		return nil, errors.New("write errors" + err.Error())
	}

	financingBytes, err = json.Marshal(&financing)

	return financingBytes, nil
}

func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	var cpFrom Company
	var cpTo Company
	var cpFromId string
	var cpToId string
	var bkFromId string
	var bkToId string
	var issueNumber int
	var tsBytes []byte
	var err error

	cpFromId = args[0]
	bkFromId = args[1]
	cpToId = args[2]
	bkToId = args[3]
	issueNumber, err = strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	cpFrom, _, err = getCompanyById(stub, cpFromId)
	if err != nil {
		return nil, errors.New("get errors")
	}
	if cpFrom.BankRest[bkFromId] < issueNumber {
		return nil, errors.New("Not enough money")
	}

	cpTo, _, err = getCompanyById(stub, cpToId)
	if err != nil {
		return nil, errors.New("get errors")
	}

	cpFrom.RestNumber -= issueNumber
	cpFrom.BankRest[bkFromId] -= issueNumber
	cpTo.RestNumber += issueNumber
	cpTo.BankRest[bkToId] += issueNumber

	err = writeCompany(stub, cpFrom)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())

	}

	err = writeCompany(stub, cpTo)
	if err != nil {
		cpFrom.RestNumber += issueNumber
		cpFrom.BankRest[bkFromId] += issueNumber
		err = writeCompany(stub, cpFrom)
		if err != nil {
			return nil, errors.New("roll down error")
		}
		return nil, errors.New("write Error" + err.Error())
	}

	transaction := Transaction{FromType: 2, FromCpID: cpFromId, FromBkID: bkFromId, ToType: 2, ToCpID: cpToId, ToBkID: bkToId, Time: time.Now().Unix(), Number: issueNumber, ID: transactionNo}
	err = writeTransaction(stub, transaction)
	if err != nil {
		return nil, errors.New("write Error" + err.Error())
	}

	tsBytes, err = json.Marshal(&transaction)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}

	transactionNo = transactionNo + 1
	return tsBytes, nil
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	if function == "getCenterBank" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		_, cbBytes, err := getCenterBank(stub)
		if err != nil {
			fmt.Println("Error get centerBank")
			return nil, err
		}
		return cbBytes, nil
	} else if function == "getCompanyById" {
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		_, cpBytes, err := getCompanyById(stub, args[0])
		if err != nil {
			fmt.Println("Error unmarshalling centerBank")
			return nil, err
		}
		return cpBytes, nil
	} else if function == "getTransactionById" {
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		_, tsBytes, err := getTransactionById(stub, args[0])
		if err != nil {
			fmt.Println("Error unmarshalling")
			return nil, err
		}
		return tsBytes, nil
	} else if function == "getTransactions" {
		if len(args) != 0 {
			return nil, errors.New("Incorrect number of arguments. Expecting 0")
		}
		tss, err := getTransactions(stub)
		if err != nil {
			fmt.Println("Error unmarshalling")
			return nil, err
		}
		tsBytes, err1 := json.Marshal(&tss)
		if err1 != nil {
			fmt.Println("Error marshalling banks")
		}
		return tsBytes, nil
	}
	return nil, nil
}

func getCenterBank(stub shim.ChaincodeStubInterface) (CenterBank, []byte, error) {
	var centerBank CenterBank
	cbBytes, err := stub.GetState("centerBank")
	if err != nil {
		fmt.Println("Error retrieving cbBytes")
	}
	err = json.Unmarshal(cbBytes, &centerBank)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}
	return centerBank, cbBytes, nil
}

func getCompanyById(stub shim.ChaincodeStubInterface, id string) (Company, []byte, error) {
	var company Company
	cpBytes, err := stub.GetState("company"+id)
	if err != nil {
		fmt.Println("Error retrieving cpBytes")
	}
	err = json.Unmarshal(cpBytes, &company)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}
	return company, cpBytes, nil
}

func getFinancingContractById(stub shim.ChaincodeStubInterface, id string) (FinancingContract, []byte, error) {
	var financing FinancingContract
	fcBytes, err := stub.GetState("financing" + id)
	if err != nil {
		fmt.Println("Error retrieving fcBytes")
	}
	err = json.Unmarshal(fcBytes, &financing)
	if err != nil {
		fmt.Println("Error unmarshalling FinancingCOntract")
	}
	return financing, fcBytes, nil
}

func getTransactionById(stub shim.ChaincodeStubInterface, id string) (Transaction, []byte, error) {
	var transaction Transaction
	tsBytes, err := stub.GetState("transaction" + id)
	if err != nil {
		fmt.Println("Error retrieving cpBytes")
	}
	err = json.Unmarshal(tsBytes, &transaction)
	if err != nil {
		fmt.Println("Error unmarshalling centerBank")
	}
	return transaction, tsBytes, nil
}

func getTransactions(stub shim.ChaincodeStubInterface) ([]Transaction, error) {
	var transactions []Transaction
	var number string
	var err error
	var transaction Transaction
	if transactionNo <= 10 {
		i := 0
		for i < transactionNo {
			number = strconv.Itoa(i)
			transaction, _, err = getTransactionById(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			transactions = append(transactions, transaction)
			i = i + 1
		}
	} else {
		i := 0
		for i < 10 {
			number = strconv.Itoa(i)
			transaction, _, err = getTransactionById(stub, number)
			if err != nil {
				return nil, errors.New("Error get detail")
			}
			transactions = append(transactions, transaction)
			i = i + 1
		}
		return transactions, nil
	}
	return nil, nil
}

func writeCenterBank(stub shim.ChaincodeStubInterface, centerBank CenterBank) error {
	cbBytes, err := json.Marshal(&centerBank)
	if err != nil {
		return err
	}
	err = stub.PutState("centerBank", cbBytes)
	if err != nil {
		return errors.New("PutState Error" + err.Error())
	}
	return nil
}

// Write the financing request to the block chain
func writeFinancingContract(stub shim.ChaincodeStubInterface, ContractId string, financing FinancingContract) error {
	financingBytes, err := json.Marshal(&financing)
	if err != nil {
		return err
	}

	err = stub.PutState("financing"+ContractId, financingBytes)
	if err != nil {
		return errors.New("PutState Error" + err.Error())
	}
	return nil
}

func writeCompany(stub shim.ChaincodeStubInterface, company Company) error {
	var companyId string
	cpBytes, err := json.Marshal(&company)
	if err != nil {
		return err
	}
	companyId = company.ID
	err = stub.PutState("company"+companyId, cpBytes)
	if err != nil {
		return errors.New("PutState Error" + err.Error())
	}
	return nil
}

func writeTransaction(stub shim.ChaincodeStubInterface, transaction Transaction) error {
	var tsId string
	tsBytes, err := json.Marshal(&transaction)
	if err != nil {
		return err
	}
	tsId = strconv.Itoa(transaction.ID)
	if err != nil {
		return errors.New("want Integer number")
	}
	err = stub.PutState("transaction"+tsId, tsBytes)
	if err != nil {
		return errors.New("PutState Error" + err.Error())
	}
	return nil
}