package transaction

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

type TxVersion2 struct{}

func txSigPubKeyToBytes(indexes [][]*big.Int) ([]byte, error) {
	n := len(indexes)
	if n == 0 {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is empty")
	}
	if n > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many rows")
	}
	m := len(indexes[0])
	if m > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many columns")
	}
	for i := 1; i < n; i += 1 {
		if len(indexes[i]) != m {
			return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is not a rectangle array")
		}
	}

	b := make([]byte, 0)
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			currentByte := indexes[i][j].Bytes()
			lengthByte := len(currentByte)
			if lengthByte > MaxSizeByte {
				return nil, errors.New("TxSigPublicKeyVer2.ToBytes: IndexesByte is too large")
			}
			b = append(b, byte(lengthByte))
			b = append(b, currentByte...)
		}
	}
	return b, nil
}

func txSigPubKeyFromBytes(b []byte) ([][]*big.Int, error) {
	if len(b) < 2 {
		return nil, errors.New("txSigPubKeyFromBytes: cannot parse length of indexes, length of input byte is too small")
	}
	n := int(b[0])
	m := int(b[1])
	offset := 2
	indexes := make([][]*big.Int, n)
	for i := 0; i < n; i += 1 {
		row := make([]*big.Int, m)
		for j := 0; j < m; j += 1 {
			if offset >= len(b) {
				return nil, errors.New("txSigPubKeyFromBytes: cannot parse byte length of index[i][j], length of input byte is too small")
			}
			byteLength := int(b[offset])
			offset += 1
			if offset+byteLength > len(b) {
				return nil, errors.New("txSigPubKeyFromBytes: cannot parse big int index[i][j], length of input byte is too small")
			}
			currentByte := b[offset : offset+byteLength]
			offset += byteLength
			row[j] = new(big.Int).SetBytes(currentByte)
		}
		indexes[i] = row
	}

	return indexes, nil
}

func generateMlsagRingWithIndexes(inp *[]*coin.InputCoin, out *[]*coin.OutputCoin, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	inputCoins := *inp
	outputCoins := *out

	// Remember which coin commitment existed in inputCoin
	// The coinCommitment is in the original format (have not changed it to ver2)
	// The reason why it must be in original format is that we will query db these commitments
	listUsableCommitments := make(map[common.Hash][]byte)
	for _, in := range inputCoins {
		usableCommitment := in.CoinDetails.GetCoinCommitment().ToBytesS()
		commitmentInHash := common.HashH(usableCommitment)
		listUsableCommitments[commitmentInHash] = usableCommitment
	}
	lenCommitment, err := statedb.GetCommitmentLength(params.stateDB, *params.tokenID, shardID)
	if err != nil {
		Logger.Log.Errorf("Getting length of commitment error %v ", err)
		return nil, nil, err
	}
	if lenCommitment == nil {
		Logger.Log.Error(errors.New("Commitments is empty"))
		return nil, nil, errors.New("Commitments is empty")
	}

	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		commitment := coin.ParseCommitmentToV2WithCoin(outputCoins[i].CoinDetails)
		outputCommitments.Add(outputCommitments, commitment)
	}

	ring := make([][]*operation.Point, ringSize)
	key := params.senderSK

	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(params.fee),
	)

	// The indexes array is for validator recheck
	indexes := make([][]*big.Int, ringSize)
	for i := 0; i < ringSize; i += 1 {
		sumInputs := new(operation.Point).Identity()
		sumInputs.Sub(sumInputs, feeCommitment)

		row := make([]*operation.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))

		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				privKey := new(operation.Scalar).FromBytesS(*key)
				row[j] = new(operation.Point).ScalarMultBase(privKey)

				coinCommitmentV2 := coin.ParseCommitmentToV2WithCoin(inputCoins[j].CoinDetails)
				sumInputs.Add(sumInputs, coinCommitmentV2)

				// Store index for validator recheck
				coinCommitmentDB := inputCoins[j].CoinDetails.GetCoinCommitment()
				commitmentBytes := coinCommitmentDB.ToBytesS()
				rowIndexes[j], err = statedb.GetCommitmentIndex(params.stateDB, *params.tokenID, commitmentBytes, shardID)
				if err != nil {
					Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, err
				}
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				for {
					index, _ := common.RandBigIntMaxRange(lenCommitment)
					rowIndexes[j] = index

					ok, err := statedb.HasCommitmentIndex(params.stateDB, *params.tokenID, index.Uint64(), shardID)
					if !ok || err != nil {
						Logger.Log.Errorf("Has commitment index error %v ", err)
						return nil, nil, err
					}
					commitment, publicKey, snd, err := statedb.GetCommitmentPublicKeyAddditionalByIndex(params.stateDB, *params.tokenID, index.Uint64(), shardID)
					if err != nil {
						Logger.Log.Errorf("Get Commitment PublicKey and Additional by index error %v ", err)
						return nil, nil, err
					}
					_, found := listUsableCommitments[common.HashH(commitment)]
					if found && (lenCommitment.Uint64() != 1 || len(inputCoins) != 1) {
						continue
					}
					row[j], err = new(operation.Point).FromBytesS(publicKey)
					if err != nil {
						fmt.Println(publicKey)
						Logger.Log.Errorf("Parsing from byte to point error %v ", err)
						return nil, nil, err
					}

					// Change commitment to v2
					commitmentBytesV2, err := coin.ParseCommitmentToV2ByBytes(
						commitment,
						publicKey,
						snd,
						shardID,
					)
					if err != nil {
						Logger.Log.Errorf("ParseCommitmentToV2ByBytes got error %v ", err)
						return nil, nil, err
					}

					temp, err := new(operation.Point).FromBytesS(commitmentBytesV2)
					if err != nil {
						Logger.Log.Errorf("commitmentBytesV2 is not byte operation.point %v ", err)
						return nil, nil, err
					}
					sumInputs.Add(sumInputs, temp)
					break
				}
			}
		}
		row = append(row, sumInputs.Sub(sumInputs, outputCommitments))
		ring[i] = row
		indexes[i] = rowIndexes
	}
	mlsagring := mlsag.NewRing(ring)
	return mlsagring, indexes, nil
}

func createPrivKeyMlsag(inp *[]*coin.InputCoin, out *[]*coin.OutputCoin, senderSK *key.PrivateKey) *[]*operation.Scalar {
	inputCoins := *inp
	outputCoins := *out

	sumRand := new(operation.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.CoinDetails.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Sub(sumRand, out.CoinDetails.GetRandomness())
	}

	sk := new(operation.Scalar).FromBytesS(*senderSK)
	privKeyMlsag := make([]*operation.Scalar, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		privKeyMlsag[i] = sk
	}
	privKeyMlsag[len(inputCoins)] = sumRand
	return &privKeyMlsag
}

func txSignatureToBytes(mlsagSignature *mlsag.MlsagSig) ([]byte, error) {
	var b []byte
	b = append(b, mlsag.MlsagPrefix)
	b = append(b, mlsagSignature.GetC().ToBytesS()...)

	r := mlsagSignature.GetR()
	n := len(r)
	if n == 0 {
		return nil, errors.New("Cannot parse byte of txSignature: length of r is 0")
	}
	if n > MaxSizeByte {
		return nil, errors.New("Cannot parse byte of txSignature: length row of r is larger than 256")
	}
	m := len(r[0])
	if m > MaxSizeByte {
		return nil, errors.New("Cannot parse byte of txSignature: length column of r is larger than 256")
	}
	for i := 1; i < m; i += 1 {
		if len(r[i]) != m {
			return nil, errors.New("Cannot parse byte of txSignature: r is not a proper rectangle")
		}
	}
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			b = append(b, r[i][j].ToBytesS()...)
		}
	}
	return b, nil
}

func txSignatureFromBytesAndKeyImages(b []byte, keyImages []*operation.Point) (*mlsag.MlsagSig, error) {
	if len(b) == 0 {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: Length of byte is 0")
	}
	if b[0] != mlsag.MlsagPrefix {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: first byte is not mlsagPrefix")
	}
	offset := 1
	if offset+operation.Ed25519KeySize > len(b) {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse C")
	}
	c := new(operation.Scalar).FromBytesS(b[offset : offset+operation.Ed25519KeySize])
	offset += operation.Ed25519KeySize
	if offset+2 > len(b) {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse n and m")
	}
	n := int(b[offset])
	m := int(b[offset+1])
	offset += 2
	r := make([][]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		row := make([]*operation.Scalar, m)
		for j := 0; j < m; j += 1 {
			if offset+operation.Ed25519KeySize > len(b) {
				return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse value of R")
			}
			currentBytes := b[offset : offset+operation.Ed25519KeySize]
			offset += operation.Ed25519KeySize
			row[j] = new(operation.Scalar).FromBytesS(currentBytes)
		}
		r[i] = row
	}
	return mlsag.NewMlsagSig(c, keyImages, r)
}

// signTx - signs tx
func signTxVer2(inp *[]*coin.InputCoin, out *[]*coin.OutputCoin, tx *Tx, params *TxPrivacyInitParams) error {
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}

	ringSize := privacy.RingSize
	if !params.hasPrivacy {
		ringSize = 1
	}

	var pi int = common.RandIntInterval(0, ringSize-1)
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)

	ring, indexes, err := generateMlsagRingWithIndexes(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}
	privKeysMlsag := *createPrivKeyMlsag(inp, out, params.senderSK)
	keyImages := mlsag.ParseKeyImages(privKeysMlsag)
	for i := 0; i < len(tx.Proof.GetInputCoins()); i += 1 {
		tx.Proof.GetInputCoins()[i].CoinDetails.SetSerialNumber(keyImages[i])
	}

	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)

	tx.sigPrivKey, err = privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		Logger.Log.Errorf("tx.SigPrivKey cannot parse arrayScalar to Bytes, error %v ", err)
		return err
	}

	tx.SigPubKey, err = txSigPubKeyToBytes(indexes)
	if err != nil {
		Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	message := tx.Proof.Bytes()
	mlsagSignature, err := sag.Sign(message)
	if err != nil {
		Logger.Log.Errorf("Cannot sign mlsagSignature, error %v ", err)
		return err
	}

	tx.Sig, err = txSignatureToBytes(mlsagSignature)
	return err
}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	fmt.Println("Proving")
	fmt.Println("Proving")
	fmt.Println("Proving")
	fmt.Println("Proving")
	outputCoins, err := parseOutputCoins(params)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputcoin, error %v ", err)
		return err
	}
	for i := 0; i < len(*outputCoins); i += 1 {
		(*outputCoins)[i].CoinDetails.SetRandomness(operation.RandomScalar())
		(*outputCoins)[i].CoinDetails.SetCoinCommitment(
			coin.ParseCommitmentToV2WithCoin((*outputCoins)[i].CoinDetails),
		)
	}
	inputCoins := &params.inputCoins

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, &params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	err = signTxVer2(inputCoins, outputCoins, tx, params)
	fmt.Println("Done Proving")
	fmt.Println("Done Proving")
	fmt.Println("Done Proving")
	fmt.Println("Done Proving")
	fmt.Println("Done Proving")
	return err
}

func (txVer2 *TxVersion2) ProveASM(tx *Tx, params *TxPrivacyInitParamsForASM) error {
	return txVer2.Prove(tx, &params.txParam)
}

func getRingFromIndexesWithDatabase(tx *Tx, indexes [][]*big.Int, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (*mlsag.Ring, error) {
	var err error
	n := len(indexes)
	m := len(indexes[0])

	if n == 0 {
		return nil, errors.New("Cannot get ring from indexes: Indexes is empty")
	}

	// This is txver2 so outputCoin should be in txver2 format
	outputCoins := tx.Proof.GetOutputCoins()
	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		commitment := outputCoins[i].CoinDetails.GetCoinCommitment()
		outputCommitments.Add(outputCommitments, commitment)
	}
	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(tx.Fee),
	)

	ring := make([][]*operation.Point, n)
	for i := 0; i < n; i += 1 {
		sumCommitment := new(operation.Point).Identity()
		sumCommitment.Sub(sumCommitment, feeCommitment)

		row := make([]*operation.Point, m+1)
		for j := 0; j < m; j += 1 {
			index := indexes[i][j]
			ok, err := statedb.HasCommitmentIndex(transactionStateDB, *tokenID, index.Uint64(), shardID)
			if !ok || err != nil {
				Logger.Log.Errorf("HasCommitmentIndex error %v ", err)
				return nil, err
			}
			commitmentByte, publicKeyByte, sndByte, err := statedb.GetCommitmentPublicKeyAddditionalByIndex(transactionStateDB, *tokenID, index.Uint64(), shardID)
			if err != nil {
				Logger.Log.Errorf("Get Commitment, PublicKey, Additional by index error %v ", err)
				return nil, err
			}
			row[j], err = new(operation.Point).FromBytesS(publicKeyByte)
			if err != nil {
				Logger.Log.Errorf("Parse bytes to mlsagRing error %v ", err)
				return nil, err
			}

			commitmentV2Byte, err := coin.ParseCommitmentToV2ByBytes(
				commitmentByte,
				publicKeyByte,
				sndByte,
				shardID,
			)
			if err != nil {
				Logger.Log.Errorf("Parsing to commitmentv2 by bytes error %v ", err)
				return nil, err
			}
			commitmentV2, err := new(operation.Point).FromBytesS(commitmentV2Byte)
			if err != nil {
				Logger.Log.Errorf("CommitmentV2 from BytesS got error %v ", err)
				return nil, err
			}
			sumCommitment.Add(sumCommitment, commitmentV2)
		}
		sumCommitment.Sub(sumCommitment, outputCommitments)
		byteCommitment := sumCommitment.ToBytesS()
		row[m], err = new(operation.Point).FromBytesS(byteCommitment)
		if err != nil {
			Logger.Log.Errorf("Getting last column commitment fromBytesS got error %v ", err)
			return nil, err
		}
		ring[i] = row
	}

	// Why???
	if isNewTransaction {
		for i := 0; i < len(outputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, outputCoins[i].CoinDetails.GetSNDerivator(), transactionStateDB); ok || err != nil {
				if err != nil {
					Logger.Log.Error(err)
				}
				Logger.Log.Errorf("snd existed: %d\n", i)
				return nil, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}
	}

	mlsagRing := mlsag.NewRing(ring)
	return mlsagRing, nil
}

// verifySigTx - verify signature on tx
func verifySigTxVer2(tx *Tx, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}
	var err error

	indexes, err := txSigPubKeyFromBytes(tx.SigPubKey)
	if err != nil {
		Logger.Log.Errorf("Error when parsing bytes to txSigPubKey: %v ", err)
		return false, err
	}
	ring, err := getRingFromIndexesWithDatabase(tx, indexes, transactionStateDB, shardID, tokenID, isNewTransaction)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].CoinDetails.GetSerialNumber()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = privacy.RandomPoint()

	txSig, err := txSignatureFromBytesAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	message := tx.Proof.Bytes()
	return mlsag.Verify(txSig, ring, message)
}

// TODO privacy
func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	fmt.Println("Verifying")
	fmt.Println("Verifying")
	fmt.Println("Verifying")
	fmt.Println("Verifying")

	var valid bool
	var err error

	tokenID, err = parseTokenID(tokenID)
	if err != nil {
		return false, err
	}

	if valid, err := verifySigTxVer2(tx, transactionStateDB, shardID, tokenID, isNewTransaction); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature ver2 with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String()))
	}

	if tx.Proof == nil {
		return true, nil
	}

	// Verify the payment proof
	var txProofV2 *privacy.ProofV2 = tx.Proof.(*privacy.ProofV2)
	fmt.Println("Done Verifying")
	fmt.Println("Done Verifying")
	fmt.Println("Done Verifying")
	fmt.Println("Done Verifying")

	valid, err = txProofV2.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, shardID, tokenID, isBatch, nil)
	if !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		Logger.Log.Error("FAILED VERIFICATION PAYMENT PROOF VER 2")
		err1, ok := err.(*privacy.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == privacy.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
				if isNewTransaction {
					return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
				} else {
					// for old txs which be get from sync block or validate new block
					if tx.LockTime <= ValidateTimeForOneoutOfManyProof {
						// only verify by sign on block because of issue #504(that mean we should pass old tx, which happen before this issue)
						return true, nil
					} else {
						return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					}
				}
			}
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}
	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}
