package notifier

import (
	"github.com/rs/zerolog/log"
	"github.com/samuel/go-zookeeper/zk"

	"github.com/amazingchow/consistent-hashing-service-provider/internal/common"
	conf "github.com/amazingchow/consistent-hashing-service-provider/internal/config"
)

type Notifier struct {
	conn *zk.Conn
}

func NewNotifier(cfg *conf.Notifier) *Notifier {
	conn, _, err := zk.Connect(cfg.ZkEndpoints, common.LeaderNotifyHeartbeat)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to connect to zookeeper cluster (%v)", cfg.ZkEndpoints)
	}
	prepare(conn)
	return &Notifier{
		conn: conn,
	}
}

func prepare(conn *zk.Conn) {
	_, err := conn.Create(common.LeaderNotifyRootPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
	if err != nil && err != zk.ErrNodeExists {
		log.Fatal().Err(err).Msgf("failed to create znode (%s)", common.LeaderNotifyRootPath)
	}
	_, err = conn.Create(common.LeaderNotifyPath, []byte(""), 0, zk.WorldACL(zk.PermAll))
	if err != nil && err != zk.ErrNodeExists {
		log.Fatal().Err(err).Msgf("failed to create znode (%s)", common.LeaderNotifyPath)
	}
}

func (nf *Notifier) SetLeader(leader string) {
	_, err := nf.conn.Set(common.LeaderNotifyPath, []byte(leader), -1)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to set value for znode (%s)", common.LeaderNotifyPath)
	}
}

func (nf *Notifier) GetLeader() string {
	v, _, err := nf.conn.Get(common.LeaderNotifyPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to get value from znode (%s)", common.LeaderNotifyPath)
	}
	return string(v)
}

func (nf *Notifier) Close() {
	nf.conn.Close()
}
