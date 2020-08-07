package metrics

const (
	Measurement      = "Measurement"
	Tag              = "Tag"
	TagValue         = "TagValue"
	MeasurementValue = "MeasurementValue"
	Time             = "Time"
	GrafanaURL       = "http://128.199.96.206:8086/write?db=mydb"
)

// Measurement
const (
	TxPoolValidated                  = "TxPoolValidated"
	TxPoolValidationDetails          = "TxPoolValidationDetails"
	TxPoolValidatedWithType          = "TxPoolValidatedWithType"
	TxPoolEntered                    = "TxPoolEntered"
	TxPoolEnteredWithType            = "TxPoolEnteredWithType"
	TxPoolAddedAfterValidation       = "TxPoolAddedAfterValidation"
	TxPoolRemoveAfterInBlock         = "TxPoolRemoveAfterInBlock"
	TxPoolRemoveAfterInBlockWithType = "TxPoolRemoveAfterInBlockWithType"
	TxPoolRemoveAfterLifeTime        = "TxPoolRemoveAfterLifeTime"
	TxAddedIntoPoolType              = "TxAddedIntoPoolType"
	TxPoolPrivacyOrNot               = "TxPoolPrivacyOrNot"
	TxEnterNetSyncSuccess            = "TxEnterNetSyncSuccess"
	PoolSize                         = "PoolSize"
	TxInOneBlock                     = "TxInOneBlock"
	TxPoolDuplicateTxs               = "DuplicateTxs"
	NumOfBlockInsertToChain          = "NumOfBlockInsertToChain"
	NumOfRoundPerBlock               = "NumOfRoundPerBlock"
	TxPoolRemovedNumber              = "TxPoolRemovedNumber"
	TxPoolRemovedTime                = "TxPoolRemovedTime"
	TxPoolRemovedTimeDetails         = "TxPoolRemovedTimeDetails"
	TxPoolTxBeginEnter               = "TxPoolTxBeginEnter"
	ProcessDiscoverPeersTime         = "ProcessDiscoverPeersTime"
	AllConnectedPeers                = "AllConnectedPeers"
	BeaconBlock                      = "BeaconBlock"
	ShardBlock                       = "ShardBlock"
	CreateNewShardBlock              = "CreateNewShardBlock"
	HandleAllMessage                 = "HandleAllMessage"
	HandleAllMessageSize             = "HandleAllMessageSize"
	HandleMessagePeerState           = "HandleMessagePeerState"
	HandleMessageBFTMsg              = "HandleMessageBFTMsg"
	HandleMessagePeerStateTime       = "HandleMessagePeerStateTime"
	HandleMessageBFTMsgTime          = "HandleMessageBFTMsgTime"
	HandleMessageGetBlockBeacon      = "HandleMessageGetBlockBeacon"
	HandleMessageGetShardToBeacon    = "HandleMessageGetShardToBeacon"
	HandleMessageGetCrossShard       = "HandleMessageGetCrossShard"
	HandleMessageGetBlockShard       = "HandleMessageGetBlockShard"
	HandleMessageShardToBeacon       = "HandleMessageShardToBeacon"
	HandleMessageCrossShard          = "HandleMessageCrossShard"
	HandleMessageShardBlock          = "HandleMessageShardBlock"
	HandleMessageBeaconBlock         = "HandleMessageBeaconBlock"
	NumberOfGoRoutine                = "NumberOfGoRoutine"
)

// tag
const (
	BlockHeightTag              = "blockheight"
	TxSizeTag                   = "txsize"
	TxSizeWithTypeTag           = "txsizewithtype"
	PoolSizeMetric              = "poolsize"
	TxTypeTag                   = "txtype"
	ValidateConditionTag        = "validatecond"
	TxPrivacyOrNotTag           = "txprivacyornot"
	ShardIDTag                  = "shardid"
	NodeIDTag                   = "node"
	TxHashTag                   = "txhash"
	FuncTag                     = "func"
	ExternalAddressTag          = "externaladdresstag"
	NewShardBlockProcessingStep = "newshardblockprocessingstep"
)

//Tag value
const (
	Beacon                                          = "beacon"
	Shard                                           = "shard"
	TxPrivacy                                       = "privacy"
	TxNormalPrivacy                                 = "normaltxprivacy"
	TxNoPrivacy                                     = "noprivacy"
	TxNormalNoPrivacy                               = "normaltxnoprivacy"
	Condition1                                      = "condition1"
	Condition2                                      = "condition2"
	Condition3                                      = "condition3"
	Condition4                                      = "condition4"
	Condition41                                     = "condition42"
	Condition42                                     = "condition42"
	Condition5                                      = "condition5"
	Condition6                                      = "condition6"
	Condition7                                      = "condition7"
	Condition8                                      = "condition8"
	Condition9                                      = "condition9"
	Condition10                                     = "condition10"
	Condition11                                     = "condition11"
	VTBITxTypeMetic                                 = "vtbitxtype"
	ReplaceTxMetic                                  = "replacetx"
	CloneShardBestStateStep                         = "cloneshardbeststatestep"
	FetchBeaconBlockStep                            = "fetchbeaconblockstep"
	GetCrossShardDataStep                           = "getcrossshardatastep"
	CreateNormalTokenTxFromCrossShardStep           = "createnormaltokentxfromcrossshardstep"
	GetTransactionForNewBlockStep                   = "gettransactionfornewblockstep"
	BuildResponseTransactionFromTxsWithMetadataStep = "buildresponsetransactionfromtxswithmetdatastep"
	ProcessInstructionFromBeaconStep                = "processinstructionfrombeaconstep"
	GenerateInstructionStep                         = "generateinstructionstep"
	BuildShardBlockHeaderEssentialStep              = "buildshardblockheaderessentialstep"
	UpdateShardBestStateStep                        = "updateshardbeststatestep"
	BuildHeaderRootHashStep                         = "buildheaderroothashstep"
	FullProcessingStep                              = "fullprocesscingstep"
)