package kafka

type Config struct {
	Seeds         []string `json:"endpoints" koanf:"endpoints"`
	ConsumerGroup string
}
