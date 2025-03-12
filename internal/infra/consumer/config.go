package consumer

type Config struct {
	Concurrency int `json:"concurrency,omitempty" koanf:"concurrency"`
}
