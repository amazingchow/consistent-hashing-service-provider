package consistenthashing

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"github.com/amazingchow/consistent-hashing-service-provider/internal/common"
	conf "github.com/amazingchow/consistent-hashing-service-provider/internal/config"
	myerror "github.com/amazingchow/consistent-hashing-service-provider/internal/error"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/hashlib"
	pb_api "github.com/amazingchow/consistent-hashing-service-provider/pb/api"
)

// Executor implements consistent-hashing algorithm inspired by
// "Consistent Hashing and Random Trees: Distributed Caching Protocols for Relieving Hot Spots on the World Wide Web".
type Executor struct {
	mu sync.RWMutex

	wID string
	cfg *conf.ConsistentHashing

	isReady atomic.Value

	hashspace map[uint32]string       // hash space, limited to 0~2^32-1
	nodes     map[string]*pb_api.Node // check-table for node
	nodesLoc  common.Uint32_n         // check-index for nodes
}

func NewExecutor(wID string, cfg *conf.ConsistentHashing) *Executor {
	return &Executor{
		wID: wID,
		cfg: cfg,

		hashspace: make(map[uint32]string),
		nodes:     make(map[string]*pb_api.Node),
		nodesLoc:  make(common.Uint32_n, 0),
	}
}

func (ch *Executor) Start() {
	ch.isReady.Store(1)
}

func (ch *Executor) Stop() {
	ch.isReady.Store(0)
}

func (ch *Executor) IsReady() bool {
	return ch.isReady.Load().(int) == 1
}

func (ch *Executor) Join(node *pb_api.Node) error {
	if !ch.IsReady() {
		return myerror.ErrNotReadyWorker
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if _, ok := ch.nodes[node.GetUuid()]; !ok {
		ch.join(node)
		log.Debug().Str("[Worker]", ch.wID).Msgf("node (uuid: %s) has joined", node.GetUuid())
		return nil
	}

	return myerror.NewNodeExistsError(node.GetUuid())
}

func (ch *Executor) join(node *pb_api.Node) {
	/*
		assume that uuid is "a.b.c.d", then N+1 nodes will be added.

		a.b.c.d#0
		a.b.c.d#1
		a.b.c.d#2
		...
		a.b.c.d#N (N == virtual nodes)
	*/
	for i := 0; i <= ch.cfg.VirReplicas; i++ {
		ch.hashspace[genLoc(node.GetUuid(), i)] = node.GetUuid()
	}
	ch.nodes[node.GetUuid()] = node
	ch.updateLoc()
}

func (ch *Executor) Leave(uuid string) error {
	if !ch.IsReady() {
		return myerror.ErrNotReadyWorker
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	/*
		assume that uuid is "a.b.c.d", then N+1 nodes will be removed.

		a.b.c.d#0
		a.b.c.d#1
		a.b.c.d#2
		...
		a.b.c.d#N (N == virtual nodes)
	*/
	if _, ok := ch.nodes[uuid]; ok {
		for i := 0; i <= ch.cfg.VirReplicas; i++ {
			delete(ch.hashspace, genLoc(uuid, i))
		}
		delete(ch.nodes, uuid)
		ch.updateLoc()
		// TODO: migrate keys to other nodes
		log.Debug().Str("[Worker]", ch.wID).Msgf("node (uuid: %s) has left", uuid)
		return nil
	}

	return myerror.NewNoNodeError(uuid)
}

func (ch *Executor) updateLoc() {
	// avoid memory allocation
	loc := ch.nodesLoc[:0]
	for k := range ch.hashspace { // k is node location
		loc = append(loc, k)
	}
	sort.Sort(loc)
	ch.nodesLoc = loc
}

func (ch *Executor) List() ([]*pb_api.Node, error) {
	if !ch.IsReady() {
		return nil, myerror.ErrNotReadyWorker
	}

	ch.mu.RLock()
	defer ch.mu.RUnlock()

	nodes := make([]*pb_api.Node, 0, len(ch.nodes))
	for _, v := range ch.nodes {
		nodes = append(nodes, v)
	}
	return nodes, nil
}

func (ch *Executor) Map(key string) (string, error) {
	if !ch.IsReady() {
		return "", myerror.ErrNotReadyWorker
	}

	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.nodes) == 0 {
		log.Warn().Str("[Worker]", ch.wID).Msgf("no node in hash space")
		return "", myerror.ErrNoAvailableNode
	}

	k := hashlib.FNV1av32(key)
	idx := sort.Search(len(ch.nodesLoc), func(i int) bool {
		return ch.nodesLoc[i] > k
	})
	if idx == len(ch.nodesLoc) {
		idx = 0
	}

	uuid := ch.hashspace[ch.nodesLoc[idx]]
	log.Debug().Str("[Worker]", ch.wID).Msgf("key (key: %s) has mapped on node (uuid: %s)", key, uuid)
	return uuid, nil
}
