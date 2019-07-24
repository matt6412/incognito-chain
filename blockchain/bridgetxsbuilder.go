package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func buildInstruction(
	metaType int,
	shardID byte,
	instStatus string,
	contentStr string,
) []string {
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		instStatus,
		contentStr,
	}
}

func (chain *BlockChain) buildInstructionsForIssuingReq(
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	instructions := [][]string{}
	rejectedInst := buildInstruction(metaType, shardID, "rejected", contentStr)
	db := chain.GetDatabase()
	issuingReqAction, err := metadata.ParseIssuingInstContent(contentStr)
	if err != nil {
		fmt.Println("WARNING: an issue occured while parsing issuing action content: ", err)
		return [][]string{}, err
	}

	issuingReq := issuingReqAction.Meta
	issuingTokenID := issuingReq.TokenID
	issuingTokenName := issuingReq.TokenName
	if !ac.CanProcessCIncToken(issuingTokenID) {
		fmt.Printf("WARNING: The issuing token (%s) was already used in the current block.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	ok, err := db.CanProcessCIncToken(issuingTokenID)
	if err != nil {
		fmt.Println("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return nil, err
	}
	if !ok {
		fmt.Printf("WARNING: The issuing token (%s) was already used in the previous blocks.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	issuingAcceptedInst := metadata.IssuingAcceptedInst{
		ShardID:         shardID,
		DepositedAmount: issuingReq.DepositedAmount,
		ReceiverAddr:    issuingReq.ReceiverAddress,
		IncTokenID:      issuingTokenID,
		IncTokenName:    issuingTokenName,
		TxReqID:         issuingReqAction.TxReqID,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while marshaling issuingAccepted instruction: ", err)
		return nil, err
	}

	ac.CBridgeTokens = append(ac.CBridgeTokens, &issuingTokenID)
	returnedInst := buildInstruction(metaType, shardID, "accepted", string(issuingAcceptedInstBytes))
	return append(instructions, returnedInst), nil
}

func (chain *BlockChain) buildInstructionsForIssuingETHReq(
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	instructions := [][]string{}
	rejectedInst := buildInstruction(metaType, shardID, "rejected", contentStr)
	db := chain.GetDatabase()
	issuingETHReqAction, err := metadata.ParseETHIssuingInstContent(contentStr)
	if err != nil {
		fmt.Println("WARNING: an issue occured while parsing issuing action content: ", err)
		return [][]string{}, err
	}
	md := issuingETHReqAction.Meta
	ethReceipt := issuingETHReqAction.ETHReceipt
	if ethReceipt == nil {
		fmt.Println("WARNING: eth receipt is null.")
		return append(instructions, rejectedInst), nil
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqETHTx := append(md.BlockHash[:], []byte(strconv.Itoa(int(md.TxIndex)))...)
	isUsedInBlock := metadata.IsETHTxHashUsedInBlock(uniqETHTx, ac.UniqETHTxsUsed)
	if isUsedInBlock {
		fmt.Println("WARNING: already issued for the hash in current block: ", uniqETHTx)
		return append(instructions, rejectedInst), nil
	}
	isIssued, err := db.IsETHTxHashIssued(uniqETHTx)
	if err != nil {
		fmt.Println("WARNING: an issue occured while checking the eth tx hash is issued or not: ", err)
		return nil, err
	}
	if isIssued {
		fmt.Println("WARNING: already issued for the hash in previous blocks: ", uniqETHTx)
		return append(instructions, rejectedInst), nil
	}

	logMap, err := metadata.PickNParseLogMapFromReceipt(ethReceipt)
	if err != nil {
		fmt.Println("WARNING: an error occured while parsing log map from receipt: ", err)
		return nil, err
	}
	if logMap == nil {
		fmt.Println("WARNING: could not find log map out from receipt")
		return append(instructions, rejectedInst), nil
	}

	// the token might be ETH/ERC20
	ethereumAddr, ok := logMap["_token"].(rCommon.Address)
	if !ok {
		fmt.Println("WARNING: could not parse eth token id from log map.")
		return append(instructions, rejectedInst), nil
	}
	ethereumToken := ethereumAddr.Bytes()
	canProcess, err := ac.CanProcessTokenPair(ethereumToken, md.IncTokenID)
	if err != nil {
		fmt.Println("WARNING: an error occured while checking it can process for token pair on the current block or not: ", err)
		return nil, err
	}
	if !canProcess {
		fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid in current block")
		return append(instructions, rejectedInst), nil
	}

	isValid, err := db.CanProcessTokenPair(ethereumToken, md.IncTokenID)
	if err != nil {
		fmt.Println("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: ", err)
		return nil, err
	}
	if !isValid {
		fmt.Println("WARNING: pair of incognito token id & ethereum's id is invalid with previous blocks")
		return append(instructions, rejectedInst), nil
	}

	addressStr, ok := logMap["_incognito_address"].(string)
	if !ok {
		fmt.Println("WARNING: could not parse incognito address from eth log map.")
		return append(instructions, rejectedInst), nil
	}
	amt, ok := logMap["_amount"].(*big.Int)
	if !ok {
		fmt.Println("WARNING: could not parse amount from eth log map.")
		return append(instructions, rejectedInst), nil
	}
	amount := uint64(0)
	if bytes.Equal(rCommon.HexToAddress(common.ETH_ADDR_STR).Bytes(), ethereumToken) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20
		amount = amt.Uint64()
	}

	issuingETHAcceptedInst := metadata.IssuingETHAcceptedInst{
		ShardID:         shardID,
		IssuingAmount:   amount,
		ReceiverAddrStr: addressStr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingETHReqAction.TxReqID,
		UniqETHTx:       uniqETHTx,
		ExternalTokenID: ethereumToken,
	}
	issuingETHAcceptedInstBytes, err := json.Marshal(issuingETHAcceptedInst)
	if err != nil {
		fmt.Println("WARNING: an error occured while marshaling issuingETHAccepted instruction: ", err)
		return nil, err
	}
	ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqETHTx)
	ac.DBridgeTokenPair[md.IncTokenID.String()] = ethereumToken

	acceptedInst := buildInstruction(metaType, shardID, "accepted", string(issuingETHAcceptedInstBytes))
	return append(instructions, acceptedInst), nil
}

func (blockgen *BlkTmplGenerator) buildIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	fmt.Println("[Centralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, err
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		return nil, err
	}

	if shardID != issuingAcceptedInst.ShardID {
		return nil, nil
	}

	issuingRes := metadata.NewIssuingResponse(
		issuingAcceptedInst.TxReqID,
		metadata.IssuingResponseMeta,
	)

	receiver := &privacy.PaymentInfo{
		Amount:         issuingAcceptedInst.DepositedAmount,
		PaymentAddress: issuingAcceptedInst.ReceiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   issuingAcceptedInst.IncTokenName,
		PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:         issuingAcceptedInst.DepositedAmount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}

	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		producerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		tokenParams,
		blockgen.chain.config.DataBase,
		issuingRes,
		false,
		false,
		shardID,
	)

	if initErr != nil {
		Logger.log.Error(initErr)
		return nil, initErr
	}
	fmt.Println("[Centralized token issuance] Create tx ok.")
	return resTx, nil
}

func (blockgen *BlkTmplGenerator) buildETHIssuanceTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	fmt.Println("[Decentralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, err
	}
	var issuingETHAcceptedInst metadata.IssuingETHAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		return nil, err
	}

	if shardID != issuingETHAcceptedInst.ShardID {
		return nil, nil
	}

	key, err := wallet.Base58CheckDeserialize(issuingETHAcceptedInst.ReceiverAddrStr)
	if err != nil {
		return nil, err
	}

	receiver := &privacy.PaymentInfo{
		Amount:         issuingETHAcceptedInst.IssuingAmount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingETHAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   common.PETHTokenName,
		// PropertySymbol: common.PETHTokenName,
		Amount:      issuingETHAcceptedInst.IssuingAmount,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}

	issuingETHRes := metadata.NewIssuingETHResponse(
		issuingETHAcceptedInst.TxReqID,
		issuingETHAcceptedInst.UniqETHTx,
		issuingETHAcceptedInst.ExternalTokenID,
		metadata.IssuingETHResponseMeta,
	)
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		producerPrivateKey,
		[]*privacy.PaymentInfo{},
		nil,
		0,
		tokenParams,
		blockgen.chain.config.DataBase,
		issuingETHRes,
		false,
		false,
		shardID,
	)

	if initErr != nil {
		return nil, initErr
	}
	fmt.Println("[Decentralized bridge token issuance] Create tx ok.")
	return resTx, nil
}
