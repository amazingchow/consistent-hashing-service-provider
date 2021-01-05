package service

import (
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	myerror "github.com/amazingchow/photon-dance-consistent-hashing/internal/error"
)

// Forwarder forwards the incoming traffic to next opened grpc connection
type Forwarder struct {
	mu       sync.Mutex
	lastAddr string
	lastConn *grpc.ClientConn
}

// NewForwarder returns a new Forwarder instance.
func NewForwarder() *Forwarder {
	return &Forwarder{}
}

// Connect creates next opened grpc connection.
func (f *Forwarder) Connect(addr string) (*grpc.ClientConn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if addr == "" {
		return nil, myerror.ErrBadAddress
	}
	if addr == f.lastAddr {
		return f.lastConn, nil
	}

	// TODO: grpc dial never blocks, it's safe to have it in mutex
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	if f.lastConn != nil {
		if err = f.lastConn.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close old connection inside forwarder")
		}
	}

	log.Info().Msgf("dial to %s to forward incoming request", addr)

	f.lastAddr = addr
	f.lastConn = conn

	return f.lastConn, nil
}

// Close closes last opened grpc connection.
func (f *Forwarder) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastAddr = ""
	if f.lastConn == nil {
		return nil
	}
	return f.lastConn.Close()
}
