package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
	"reflect"
	"errors"
)

// PortalUserRegister - User register porting public tokens
type PortalUserRegister struct {
	MetadataBase
	UniqueRegisterId string //
	IncogAddressStr string
	PTokenId string
	RegisterAmount uint64
}

type PortalUserRegisterAction struct {
	Meta PortalUserRegister
	TxReqID common.Hash
	ShardID byte
}

func NewPortalUserRegister(uniqueRegisterId string , incogAddressStr string, pTokenId string, registerAmount uint64, metaType int) (*PortalUserRegister, error){
	metadataBase := MetadataBase{
		Type: metaType,
	}

	portalUserRegisterMeta := &PortalUserRegister {
		UniqueRegisterId: uniqueRegisterId,
		IncogAddressStr: incogAddressStr,
		PTokenId: pTokenId,
		RegisterAmount: registerAmount,
	}

	portalUserRegisterMeta.MetadataBase = metadataBase

	return portalUserRegisterMeta, nil
}

//todo
func (portalUserRegister PortalUserRegister) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}


func (portalUserRegister PortalUserRegister) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
		return true, true, nil
	}

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(portalUserRegister.IncogAddressStr)
	if err != nil {
		return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("ContributorAddressStr incorrect"))
	}

	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}
	if !bytes.Equal(txr.GetSigPubKey()[:], incogAddr.Pk[:]) {
		return false, false, errors.New("custodian incognito address is not signer tx")
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// check burning tx
	if !txr.IsCoinsBurning(bcr) {
		return false, false, errors.New("must send coin to burning address")
	}

	//todo: verify porting fees

	// validate amount register
	if portalUserRegister.RegisterAmount == 0 {
		return false, false, errors.New("register amount should be larger than 0")
	}

	if portalUserRegister.RegisterAmount != txr.CalculateTxValue() {
		return false, false, errors.New("register amount should be equal to the tx value")
	}

	return true, true, nil
}

func (portalUserRegister PortalUserRegister) ValidateMetadataByItself() bool {
	return portalUserRegister.Type == PortalUserRegisterMeta
}

func (portalUserRegister PortalUserRegister) Hash() *common.Hash {
	record := portalUserRegister.MetadataBase.Hash().String()
	record += portalUserRegister.IncogAddressStr
	//todo:
	//record += custodianDeposit.RemoteAddresses
	record += strconv.FormatUint(portalUserRegister.RegisterAmount, 10)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (portalUserRegister *PortalUserRegister) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := PortalUserRegisterAction{
		Meta:    *portalUserRegister,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalUserRegisterMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (portalUserRegister *PortalUserRegister) CalculateSize() uint64 {
	return calculateSize(portalUserRegister)
}


