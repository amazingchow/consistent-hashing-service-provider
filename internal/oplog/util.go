package oplog

import (
	"os"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

func GetKafkaAccessEnv(cfg *sarama.Config) {
	uname := os.Getenv("KAFKA_USERNAME")
	upass := os.Getenv("KAFKA_PASSWORD")
	if uname == "" || upass == "" {
		log.Info().Msg("access kafka without SASL settings")
		return
	}
	cfg.Net.SASL.Enable = true
	cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	cfg.Net.SASL.User = uname
	cfg.Net.SASL.Password = upass
	cfg.Net.SASL.Version = sarama.SASLHandshakeV1
}
