package common

import "time"

const (
	LeaderNotifyRootPath  = "/consistent-hashing-service-provider"
	LeaderNotifyPath      = "/consistent-hashing-service-provider/leader"
	LeaderNotifyHeartbeat = 10 * time.Second
	ForwardTimeout        = 10 * time.Second
)
