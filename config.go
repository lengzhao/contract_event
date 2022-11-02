package contractevent

import _ "embed"

type SubscriptionConf struct {
	Alias      string            `yaml:"alias"`
	Contract   []string          `yaml:"contract"`
	ABIFile    string            `yaml:"abi_file"`
	EventName  string            `yaml:"event_name"`
	Filter     map[string]string `yaml:"filter"`
	StartBlock uint64            `yaml:"start_block"`
	WebHook    string            `yaml:"web_hook"`
}

type DBConf struct {
	Engine   string `yaml:"engine,omitempty"`
	DSN      string `yaml:"dsn,omitempty"`
	LogLevel int    `yaml:"log_level,omitempty"`
}

type ChainConfig struct {
	RPCNode    string `yaml:"rpc_node,omitempty"`
	DelayBlock uint64 `yaml:"delay_block,omitempty"`
}

type ServerConfig struct {
	Port       int    `yaml:"port,omitempty"`
	PrefixPath string `yaml:"prefix_path,omitempty"`
}

type Config struct {
	Chain ChainConfig        `yaml:"chain,omitempty"`
	DB    DBConf             `yaml:"db,omitempty"`
	Subs  []SubscriptionConf `yaml:"subscriptions,omitempty"`
	Http  ServerConfig       `yaml:"http,omitempty"`
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
