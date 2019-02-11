package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

const Decimals = uint64(10000) // Each float number is multiplied by this value to store as uint64

type LoanPayment struct {
	LoanID []byte
	MetadataBase
}

func NewLoanPayment(data map[string]interface{}) (Metadata, error) {
	result := LoanPayment{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s
	result.Type = LoanPaymentMeta
	return &result, nil
}

func (lp *LoanPayment) Hash() *common.Hash {
	record := string(lp.LoanID)

	// final hash
	record += lp.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lp *LoanPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Start validating LoanPayment tx with blockchain!!!")
	// Check if loan is withdrawed
	_, _, _, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return false, err
	}

	// Check loan payment
	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	burnPk := keyWalletBurningAdd.KeySet.PaymentAddress.Pk
	unique, receiver, amount := txr.GetUniqueReceiver()
	fmt.Printf("unique, receiver, amount: %v, %x, %v\n", unique, receiver, amount)
	if !unique || !bytes.Equal(receiver, burnPk) {
		return false, errors.New("Loan payment must be sent to burn address")
	}

	return true, nil
}

func (lp *LoanPayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	proof := txr.GetProof()
	if proof == nil || len(proof.InputCoins) < 1 || len(proof.OutputCoins) < 1 {
		return false, false, errors.Errorf("Loan payment must send Constant")
	}
	return true, true, nil // continue checking for fee
}

func (lp *LoanPayment) ValidateMetadataByItself() bool {
	return true
}

func GetTotalInterest(principle, interest, interestRate, maturity, deadline, currentHeight uint64) uint64 {
	totalInterest := uint64(0)
	if currentHeight >= deadline {
		perTerm := GetInterestPerTerm(principle, interestRate)
		totalInterest = interest
		if perTerm > 0 {
			totalInterest += (currentHeight - deadline) / maturity * perTerm
		}
	}
	return totalInterest
}

func GetTotalDebt(principle, interest, interestRate, maturity, deadline, currentHeight uint64) uint64 {
	return principle + GetTotalInterest(principle, interest, interestRate, maturity, deadline, currentHeight)
}

func GetInterestPerTerm(principle, interestRate uint64) uint64 {
	return principle * interestRate / Decimals
}

func (lp *LoanPayment) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	amount, err := lp.calculateInterestPaid(txr, bcr)
	if err != nil {
		return [][]string{}, err
	}
	lrActionValue := getLoanPaymentActionValue(txr.Hash(), amount)
	lrAction := []string{strconv.Itoa(LoanPaymentMeta), lrActionValue}
	return [][]string{lrAction}, nil
}

func getLoanPaymentActionValue(txHash *common.Hash, amount uint64) string {
	return strings.Join([]string{txHash.String(), strconv.FormatUint(amount, 10)}, actionValueSep)
}

func ParseLoanPaymentActionValue(values string) (*common.Hash, uint64, error) {
	s := strings.Split(values, actionValueSep)
	if len(s) != 2 {
		return nil, 0, errors.Errorf("LoanPayment value invalid")
	}
	txHash, errHash := common.NewHashFromStr(s[0])
	amount, errAmount := strconv.ParseUint(s[1], 10, 64)
	errSaver := &ErrorSaver{}
	if errSaver.Save(errHash, errAmount) != nil {
		return nil, 0, errSaver.Get()
	}
	return txHash, amount, nil
}

func (lp *LoanPayment) calculateInterestPaid(tx Transaction, bcr BlockchainRetriever) (uint64, error) {
	principle, interest, deadline, err := bcr.GetLoanPayment(lp.LoanID)
	if err != nil {
		return 0, err
	}

	// Get loan params
	requestMeta, err := bcr.GetLoanRequestMeta(lp.LoanID)
	if err != nil {
		return 0, err
	}

	// Only keep interest
	_, _, amount := tx.GetUniqueReceiver() // Receiver is unique and is burn address
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	totalInterest := GetTotalInterest(
		principle,
		interest,
		requestMeta.Params.InterestRate,
		requestMeta.Params.Maturity,
		deadline,
		bcr.GetChainHeight(shardID),
	)
	interestPaid := amount
	if amount > totalInterest {
		interestPaid = totalInterest
	}
	return interestPaid, nil
}
