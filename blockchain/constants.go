package blockchain

// constant for network
const (
	MAINNET             = 0xd9b4bef9
	MAINNET_NAME        = "mainnet"
	MAINET_DEFAULT_PORT = "9333"
)

// constant for genesis block
const (
	GENESIS_BLOCK_REWARD = 1000000000
	//GENESIS_BLOCK_PUBKEY_ADDR  = "mgnUx4Ah4VBvtaL7U1VXkmRjKUk3h8pbst"
	GENESIS_BLOCK_PAYMENT_ADDR = "12S2mXUdD37GbP6XqL3VyLFRn9BoouYKsF1gvBJC1gzcyLze9tvcHRv27mpEB77aoqArrVMSxb1xtKasNmgEgY8maJo3YwZg1ZoA7Tp"
	// readonly-key 13heMgrXxqBfc6m1TQDrp2EMhsE11vXqyTDWB9xn6e7rtiw3YeD1GEnXaPdhK9VV9BaDcihdaQTpVgsdczsYB24bjDQGWFcu9MXEvq8
	// pri-key 1111111f2rTmz8gBKi8bBs3NWkoq6dd4mN17CrURn3JoFcTtxeejgKcdGXtpA24MJXhUHWkiHNFUhRDP5Poa6FEcW62kmvsar198JLnqwX
)

// global variables for genesis blok
var (
	GENESIS_BLOCK_ANCHOR            = [32]byte{1}
	GENESIS_BLOCK_NULLIFIERS        = []string{"88d35350b1846ecc34d6d04a10355ad9a8e1252e9d7f3af130186b4faf1a9832", "286b563fc45b7d5b9f929fb2c2766382a9126483d8d64c9b0197d049d4e89bf7"}
	GENESIS_BLOCK_COMMITMENTS       = []string{"9e83d5ca518578d5c512e8a41c1df6b6e21b1783528cad0d86b4a429a01bcb5d", "88f21ba21506a297797cc599bc2209f38a95d3f39614b89f2fb299854e8407d2"}
	GENESIS_BLOCK_OUTPUT_R1         = [32]byte{1}
	GENESIS_BLOCK_OUTPUT_R2         = [32]byte{2}
	GENESIS_BLOCK_OUTPUT_R          = [][]byte{GENESIS_BLOCK_OUTPUT_R1[:], GENESIS_BLOCK_OUTPUT_R2[:]}
	GENESIS_BLOCK_SEED              = [32]byte{1}
	GENESIS_BLOCK_PHI               = [32]byte{1}
	GENESIS_BLOCK_JSPUBKEY          = "8a8ae7ff31597a4d87be0780a5c887c990c2965f454740dfc5b4177e900104c2"
	GENESIS_BLOCK_EPHEMERAL_PRIVKEY = [32]byte{1}
	GENESIS_BLOCK_EPHEMERAL_PUBKEY  = "2fe57da347cd62431528daac5fbb290730fff684afc4cfc2ed90995f58cb3b74"
	GENESIS_BLOCK_ENCRYPTED_DATA    = []string{"33714a76425a61502d32333747564b6f7a415f444a584c61575a646239704241574b503052666d6d326151714e6333504f504769354f6d417169634a39423370766c7072637858776e4a61782d5f666958684350767a68634a74744c385a45366b676333746d543865363353536e43544d37537165554272794b694a706534526144595f4c385143634361545277454a356e5062616d624a4e5950315f3573695a443656324568534e59646e6c354b31627a7551476733502d73465375456c6e376e434878466138753976684f62397648657a526777", "757a426b373438545930317044705645735f31546c477545466b727147316f44454f6c554744416c37413853664b346635344244767a495f46446f70444b30432d75555f574b75336a4d414e706b4b75736147755776713252356f734961717a7658576c4a374e36764b42705945506a483270544f593969483862734848315a50567a427357366a65414e34304a3053785a51667a473449775535634b32684f7148576c5f484f4d37763176504639427a7a7241516c7446512d327472414c30"}
	GENESIS_BLOCK_G_A               = "00156979fc20b67f8af83ff2b6267b561d2cbaef336d9dc4ae2cc686045f16e210"
	GENESIS_BLOCK_G_APrime          = "0129d18934ee3dbd94d63d378059bdae552a116c995d5c7828076652cbac773572"
	GENESIS_BLOCK_G_B               = "0106e11bfcb22cc4965d226777c9548763e17b3d684f6ccdfc82ddbb0a124be7d93af796e3395758ee4f45d035bbabd1e46c869d1ade9af87f62cd7a1706443209"
	GENESIS_BLOCK_G_BPrime          = "000796bfd9e224611238659c1288be48256ebb303982b1c15cdf5e45315e9bab5b"
	GENESIS_BLOCK_G_C               = "0004b30e804f5842bcb28437f03b1d86c0c22cc911e1f3c7de3ad169c505fd2737"
	GENESIS_BLOCK_G_CPrime          = "000e92cd29690c4405df68b52544ee98a702b342a0c06d029f295584fc77ce4c72"
	GENESIS_BLOCK_G_K               = "002ac5697c22ce57b0647951cf205de62611fcf68fb5c10743864545259d8d4f1f"
	GENESIS_BLOCK_G_H               = "0003ec3613bac9c6935e4236571279b5d2f69b26043e30037be3f6ea1dfc7ecac8"
	GENESIS_BLOCK_VMACS             = []string{"5264e44bd87cc4d555d57069f53990e9237c10d91f32e2b0c3e5ea54a9d4c7cb", "9f1187c5cf2e999904e43595ebcce0cfde7f022747b823e7134fd389c1e5a5ad"}
)
