package config

type Node struct {
	GRPCEndpoint string             `json:"grpc_endpoint"` // grpc服务地址
	CH           *ConsistentHashing `json:"ch"`
	Oplogger     *Oplogger          `json:"oplogger"`
	Notifier     *Notifier          `json:"notifier"`
	IsPrimary    bool               `json:"is_primary"`
}

type ConsistentHashing struct {
	VirReplicas int `json:"vir_replicas"` // 虚拟节点数目
}

type Oplogger struct {
	Enable             bool     `json:"enable"`
	KafkaTopic         string   `json:"kafka_topic"`
	KafkaBrokers       []string `json:"kafka_brokers"`
	KafkaConsumerGroup string   `json:"kafka_consumer_group"`
}

type Notifier struct {
	ZkEndpoints []string `json:"zk_endpoints"` // zookeeper服务地址列表
}
