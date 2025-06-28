package types

type EventMessage struct {
	Topic string
	Msg   any
}

type ProducerConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers" json:"bootstrap_servers"`
	Acks             string `yaml:"acks" json:"acks"`
	ClientID         string `yaml:"client_id" json:"client_id"`
}

type ConsumerConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers" json:"bootstrap_servers"`
	ClientID         string `yaml:"client_id" json:"client_id"`
	AutoOffsetReset  string `yaml:"auto_offset_reset" json:"auto_offset_reset"`
	EnableAutoCommit bool   `yaml:"enable_autocommit" json:"enable_autocommit"`
	SessionTimeoutMs string `yaml:"session_timeout_ms" json:"session_timeout_ms"`
	GroupID          string `yaml:"group_id" json:"group_id"`
}

type WebapiConfig struct {
	Listen string `yaml:"listen" json:"listen"`
}

type CommonConfig struct {
	EventsQueueSize int `yaml:"events_queue_size" json:"vents_queue_sizeient"`
}

type AppConfig struct {
	Topics         []string       `yaml:"topics" json:"topics"`
	ConsumerConfig ConsumerConfig `yaml:"consumer" json:"consumer"`
	ProducerConfig ProducerConfig `yaml:"producer" json:"producer"`
	WebapiConfig   WebapiConfig   `yaml:"webapi" json:"webapi"`
	CommonConfig   CommonConfig   `yaml:"common" json:"common"`
}
