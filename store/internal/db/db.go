package db

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DB struct {
	dbConn *sql.DB
}

func New(host, port, user, pass, dbname string) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%w, error in opening database connection", err)
	}

	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("%w, db ping failed", err)
	}

	return &DB{dbConn: dbConn}, nil
}

func (d DB) ReserveFood(ctx context.Context, foodId, maxReservationTimeSec int32) (packetId int32, err error) {
	var availablePacketId int32

	tx, err := d.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%w; error in starting transaction", err)
	}

	defer tx.Rollback()

	getAvailableFoodPacket := fmt.Sprintf(`select id
from store.packet
where food_id = $1 and order_id is null
  and case when reserved_at is null then true else reserved_at < (now() - interval '%d' second) end
limit 1`, maxReservationTimeSec)
	if err := tx.QueryRowContext(ctx, getAvailableFoodPacket, foodId).Scan(&availablePacketId); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no food packets available")
		}
		return 0, fmt.Errorf("%w; error in getting available food packets", err)
	}

	if _, err := tx.ExecContext(ctx, "update store.packet set reserved_at = now() where id = $1", availablePacketId); err != nil {
		return 0, fmt.Errorf("%w; error in resvering the food packet", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%w; error in commiting the transaction", err)
	}

	return availablePacketId, nil
}

func (d DB) BookFoodPacket(ctx context.Context, packetId, orderId int32) error {
	var oid sql.NullInt32
	row := d.dbConn.QueryRowContext(ctx, `select order_id from store.packet where id = $1`, packetId)
	if err := row.Scan(&oid); err != nil {
		if err != sql.ErrNoRows {
			return status.Error(codes.NotFound, "food packet id not found")
		}
		return fmt.Errorf("%w; no such packet id exists", err)
	}

	if oid.Valid {
		if oid.Int32 != orderId {
			return fmt.Errorf("food packet has been already booked for another order")
		}
		// it's possible that the server has retried the same request due to network failure or something else
		return nil
	}

	if _, err := d.dbConn.ExecContext(ctx, `update store.packet set order_id = $1 where id = $2`, orderId, packetId); err != nil {
		return fmt.Errorf("%w; error in updating the food packet with order")
	}

	return nil
}
