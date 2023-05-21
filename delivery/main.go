package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Ajithkumarsekar/two-phase-commit/contracts/gen/delivery"
	"github.com/Ajithkumarsekar/two-phase-commit/delivery/internal/db"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxReservationTimeSeconds = 10

type Server struct {
	dbOp *db.DB
	delivery.UnimplementedDeliveryServiceServer
}

func (s Server) ReserveAgent(ctx context.Context, _ *emptypb.Empty) (*delivery.ReserveAgentResponse, error) {
	agentId, err := s.dbOp.ReserveADeliveryAgent(ctx, maxReservationTimeSeconds)
	if err != nil {
		return nil, fmt.Errorf("%w, error in reserving a delivery agent", err)
	}

	return &delivery.ReserveAgentResponse{AgentId: agentId}, nil
}

func (s Server) BookAgent(ctx context.Context, request *delivery.BookAgentRequest) (*emptypb.Empty, error) {
	if err := s.dbOp.BookTheAgent(ctx, request.AgentId, request.OrderId); err != nil {
		return nil, fmt.Errorf("%w; error in booking the agent", err)
	}

	return &emptypb.Empty{}, nil
}

func main() {
	dbOpp, err := db.New("127.0.0.1", "13454", "postgres", "postgres", "postgres")
	if err != nil {
		log.Fatalf("failed to int dbOpp connection %v", err)
	}
	fmt.Printf("database connection initialized")

	listen, err := net.Listen("tcp", ":12001")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	delivery.RegisterDeliveryServiceServer(s, &Server{
		dbOp: dbOpp,
	})
	log.Printf("Delivery service starting at port 12001")
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve %v", err)
	}

}
