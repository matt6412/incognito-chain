package wire

import (
	"encoding/json"
)

type MessageInvalidBlock struct {
	Reason       string //the reason it's invalid could be in
	BlockHash    string
	ChainID      byte
	Validator    string
	ValidatorSig string
}

func (self MessageInvalidBlock) MessageType() string {
	return CmdInvalidBlock
}

func (self MessageInvalidBlock) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (self MessageInvalidBlock) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(self)
	return jsonBytes, err
}

func (self MessageInvalidBlock) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), self)
	return err
}
