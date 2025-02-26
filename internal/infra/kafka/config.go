package kafka

type Config struct {
	Seeds         []string `json:"endpoints,omitempty"      koanf:"endpoints"`
	ConsumerGroup string   `json:"consumer_group,omitempty" koanf:"consumer_group"`
}
