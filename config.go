package contractevent

import _ "embed"

type SubscriptionConf struct {
	Alias      string            `yaml:"alias"`
	Contract   []string          `yaml:"contract"`
	ABIFile    string            `yaml:"abi_file"`
	EventName  string            `yaml:"event_name"`
	Filter     map[string]string `yaml:"filter"`
	StartBlock uint64            `yaml:"start_block"`
}

const (
	ABIERC20   = "erc20"
	ABIERC721  = "erc721"
	ABIERC1155 = "erc1155"
)

//go:embed erc20.json
var abiERC20 []byte

//go:embed erc721.json
var abiERC721 []byte

//go:embed erc1155.json
var abiERC1155 []byte

var abis map[string][]byte = map[string][]byte{
	ABIERC20:   abiERC20,
	ABIERC721:  abiERC721,
	ABIERC1155: abiERC1155,
}

func GetABIData(name string) []byte {
	return abis[name]
}
