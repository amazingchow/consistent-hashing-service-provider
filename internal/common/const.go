package common

import "time"

const (
	LeaderNotifyRootPath  = "/photon-dance-consistent-hashing"
	LeaderNotifyPath      = "/photon-dance-consistent-hashing/leader"
	LeaderNotifyHeartbeat = 10 * time.Second
	ForwardTimeout        = 10 * time.Second
)
