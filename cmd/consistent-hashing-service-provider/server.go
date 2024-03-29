package main

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	consistenthashing "github.com/amazingchow/consistent-hashing-service-provider/internal/ch"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/common"
	conf "github.com/amazingchow/consistent-hashing-service-provider/internal/config"
	myerror "github.com/amazingchow/consistent-hashing-service-provider/internal/error"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/notifier"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/oplog"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/service"
	pb_api "github.com/amazingchow/consistent-hashing-service-provider/pb/api"
)

type consistentHashingServiceServer struct {
	nID       string
	isLeader  bool
	notifier  *notifier.Notifier
	executor  *consistenthashing.Executor
	oplogger  *oplog.OpLogger
	forwarder *service.Forwarder
}

func newConsistentHashingServiceServer(ctx context.Context, nID string, cfg *conf.Node) *consistentHashingServiceServer {
	executor := consistenthashing.NewExecutor(nID, cfg.CH)
	go executor.Start()
	notifier := notifier.NewNotifier(cfg.Notifier)

	var oplogger *oplog.OpLogger
	if cfg.IsPrimary {
		notifier.SetLeader(nID)
		if cfg.Oplogger.Enable {
			oplogger = oplog.NewOpLogger(ctx, cfg.Oplogger.KafkaTopic, executor)
			oplogger.SetUpProducer(cfg.Oplogger.KafkaBrokers)
		}
	} else {
		if cfg.Oplogger.Enable {
			oplogger = oplog.NewOpLogger(ctx, cfg.Oplogger.KafkaTopic, executor)
			oplogger.SetUpConsumer(cfg.Oplogger.KafkaBrokers, cfg.Oplogger.KafkaConsumerGroup)
		}
	}

	return &consistentHashingServiceServer{
		nID:       nID,
		isLeader:  cfg.IsPrimary,
		notifier:  notifier,
		executor:  executor,
		oplogger:  oplogger,
		forwarder: service.NewForwarder(),
	}
}

func (srv *consistentHashingServiceServer) close() {
	srv.notifier.Close()
	srv.oplogger.Close()
	srv.executor.Stop()
	srv.forwarder.Close()
}

func (srv *consistentHashingServiceServer) Add(ctx context.Context, req *pb_api.AddRequest) (*pb_api.AddResponse, error) {
	if req.GetNode().GetUuid() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty shard uuid")
	}

	var err error
	if !srv.isLeader {
		var cli pb_api.ConsistentHashingServiceClient
		if cli, err = srv.forwardToLeader(); err == nil && cli != nil {
			tctx, cancel := context.WithTimeout(ctx, common.ForwardTimeout)
			defer cancel()
			return cli.Add(tctx, req)
		}
	}
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}

	err = srv.executor.Join(req.GetNode())
	if err == nil {
		// Leader节点添加OpLog
		if srv.oplogger != nil {
			bytes, _ := proto.Marshal(req.GetNode())
			srv.oplogger.SyncAddOpLog(&pb_api.OpLogEntry{
				OperationType: pb_api.OperationType_OPERATION_TYPE_ADD,
				Payload:       bytes,
			})
		}
		return &pb_api.AddResponse{}, nil
	}
	err = convertError(ctx, err)
	return nil, err
}

func (srv *consistentHashingServiceServer) AddN(ctx context.Context, req *pb_api.AddNRequest) (*pb_api.AddNResponse, error) {
	if len(req.GetNodes()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no shard")
	}
	for _, shard := range req.GetNodes() {
		if shard.GetUuid() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "empty shard uuid")
		}
	}

	var err error
	if !srv.isLeader {
		var cli pb_api.ConsistentHashingServiceClient
		if cli, err = srv.forwardToLeader(); err == nil && cli != nil {
			tctx, cancel := context.WithTimeout(ctx, common.ForwardTimeout)
			defer cancel()
			return cli.AddN(tctx, req)
		}
	}
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(req.GetNodes()); i++ {
		if err = srv.executor.Join(req.GetNodes()[i]); err == nil {
			// Leader节点添加OpLog
			if srv.oplogger != nil {
				bytes, _ := proto.Marshal(req.GetNodes()[i])
				srv.oplogger.SyncAddOpLog(&pb_api.OpLogEntry{
					OperationType: pb_api.OperationType_OPERATION_TYPE_ADD,
					Payload:       bytes,
				})
			}
		}
	}
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}
	return &pb_api.AddNResponse{}, nil
}

func (srv *consistentHashingServiceServer) Delete(ctx context.Context, req *pb_api.DeleteRequest) (*pb_api.DeleteResponse, error) {
	if req.GetUuid() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty shard uuid")
	}

	var err error
	if !srv.isLeader {
		var cli pb_api.ConsistentHashingServiceClient
		if cli, err = srv.forwardToLeader(); err == nil && cli != nil {
			tctx, cancel := context.WithTimeout(ctx, common.ForwardTimeout)
			defer cancel()
			return cli.Delete(tctx, req)
		}
	}
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}

	err = srv.executor.Leave(req.GetUuid())
	if err == nil {
		// Leader节点添加OpLog
		if srv.oplogger != nil {
			srv.oplogger.SyncAddOpLog(&pb_api.OpLogEntry{
				OperationType: pb_api.OperationType_OPERATION_TYPE_REMOVE,
				Payload:       []byte(req.GetUuid()),
			})
		}
		return &pb_api.DeleteResponse{}, nil
	}
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}
	return &pb_api.DeleteResponse{}, nil
}

func (srv *consistentHashingServiceServer) List(ctx context.Context, req *pb_api.ListRequest) (*pb_api.ListResponse, error) {
	nodes, err := srv.executor.List()
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}
	return &pb_api.ListResponse{Nodes: nodes}, nil
}

func (srv *consistentHashingServiceServer) MapKey(ctx context.Context, req *pb_api.MapKeyRequest) (*pb_api.MapKeyResponse, error) {
	if req.GetKey().GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty key")
	}

	uuid, err := srv.executor.Map(req.GetKey().GetName())
	err = convertError(ctx, err)
	if err != nil {
		return nil, err
	}
	return &pb_api.MapKeyResponse{Key: &pb_api.Key{NodeUuid: uuid}}, nil
}

func (srv *consistentHashingServiceServer) forwardToLeader() (pb_api.ConsistentHashingServiceClient, error) {
	leader := srv.notifier.GetLeader()
	conn, err := srv.forwarder.Connect(leader)
	if err != nil {
		return nil, myerror.ErrNoAvailableLeader
	}
	return pb_api.NewConsistentHashingServiceClient(conn), nil
}

func convertError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case myerror.ErrNoAvailableLeader:
		{
			return status.Errorf(codes.Unavailable, err.Error())
		}
	case myerror.ErrNotReadyWorker:
		{
			return status.Errorf(codes.Unavailable, err.Error())
		}
	case myerror.ErrNoAvailableNode:
		{
			return status.Errorf(codes.Unavailable, err.Error())
		}
	default:
		{

			if _, ok := err.(*myerror.NoNodeError); ok {
				return status.Errorf(codes.NotFound, err.Error())
			}
			if _, ok := err.(*myerror.NodeExistsError); ok {
				return status.Errorf(codes.AlreadyExists, err.Error())
			}
			return nil
		}
	}
}
