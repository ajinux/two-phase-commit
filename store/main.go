package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Ajithkumarsekar/two-phase-commit/contracts/gen/store"
	"github.com/Ajithkumarsekar/two-phase-commit/store/internal/db"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxReservationTimeSeconds = 10

type Server struct {
	dbOp *db.DB
	store.UnimplementedStoreServiceServer
}

func (s Server) ReserveFood(ctx context.Context, request *store.ReserveFoodRequest) (*store.ReserveFoodResponse, error) {
	packetId, err := s.dbOp.ReserveFood(ctx, request.FoodId, maxReservationTimeSeconds)
	if err != nil {
		return nil, fmt.Errorf("%w; error in reserving the food", err)
	}

	return &store.ReserveFoodResponse{PacketId: packetId}, nil
}

func (s Server) BookFood(ctx context.Context, request *store.BookFoodRequest) (*emptypb.Empty, error) {
	if err := s.dbOp.BookFoodPacket(ctx, request.PacketId, request.OrderId); err != nil {
		return nil, fmt.Errorf("%w; error in booking the food packet", err)
	}

	return &emptypb.Empty{}, nil
}

func main() {
	dbOpp, err := db.New("127.0.0.1", "13454", "postgres", "postgres", "postgres")
	if err != nil {
		log.Fatalf("failed to int dbOpp connection %v", err)
	}
	fmt.Printf("database connection initialized")

	listen, err := net.Listen("tcp", ":12002")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	store.RegisterStoreServiceServer(s, &Server{
		dbOp: dbOpp,
	})
	log.Printf("Store service starting at port 12002")
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve %v", err)
	}

}
