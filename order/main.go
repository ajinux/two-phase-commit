package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Ajithkumarsekar/two-phase-commit/contracts/gen/delivery"
	"github.com/Ajithkumarsekar/two-phase-commit/contracts/gen/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderService struct {
	deliveryClient delivery.DeliveryServiceClient
	storeClient    store.StoreServiceClient
}

func (o OrderService) placeFoodOrder(ctx context.Context, foodId int32, orderId int32) error {
	// reserve/prepare phase
	reserveFoodResp, err := o.storeClient.ReserveFood(ctx, &store.ReserveFoodRequest{FoodId: foodId})
	if err != nil {
		return fmt.Errorf("%w; error in reserving food", err)
	}
	log.Printf("reserved food packet : %d", reserveFoodResp.PacketId)
	reserverAgentResp, err := o.deliveryClient.ReserveAgent(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("%w; error in reserving delivery agent", err)
	}
	log.Printf("reserved delivery agent : %d", reserverAgentResp.AgentId)

	// commit phase
	if _, err := o.storeClient.BookFood(ctx, &store.BookFoodRequest{
		PacketId: reserveFoodResp.PacketId,
		OrderId:  orderId,
	}); err != nil {
		// perhaps add some retry mechanism
		return fmt.Errorf("%w; error in booking food", err)
	}
	log.Print("booked food packet")

	if _, err := o.deliveryClient.BookAgent(ctx, &delivery.BookAgentRequest{
		AgentId: reserverAgentResp.AgentId,
		OrderId: orderId,
	}); err != nil {
		return fmt.Errorf("%w; error in booking the agent", err)
	}
	log.Print("booked delivery agent")

	return nil
}

func main() {
	ctx := context.Background()
	deliveryConn, err := grpc.Dial("localhost:12001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to delivery service %v", err)
	}
	defer deliveryConn.Close()
	deliveryClient := delivery.NewDeliveryServiceClient(deliveryConn)

	storeConn, err := grpc.Dial("localhost:12002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to store service %v", err)
	}
	defer storeConn.Close()
	storeClient := store.NewStoreServiceClient(storeConn)

	orderService := &OrderService{
		deliveryClient: deliveryClient,
		storeClient:    storeClient,
	}

	orderId := int32(1)
	for foodId := int32(1); foodId < 5; foodId++ {
		for orders := 1; orders < 3; orders++ {
			log.Printf("ordering food id %d for order id %d", foodId, orderId)
			if err := orderService.placeFoodOrder(ctx, foodId, orderId); err != nil {
				log.Printf("%v", err)
			}
			log.Printf("\n\n")
			orderId = orderId + 1
		}
	}
}
