package contractevent

type SubscriptionConf struct {
	Alias      string            `yaml:"alias"`
	Contract   string            `yaml:"contract"`
	ABIFile    string            `yaml:"abi_file"`
	EventName  string            `yaml:"event_name"`
	Filter     map[string]string `yaml:"filter"`
	StartBlock uint64            `yaml:"start_block"`
}
