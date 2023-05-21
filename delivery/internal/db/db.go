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

func (d *DB) ReserveADeliveryAgent(ctx context.Context, maxReserveTimeSec int) (int32, error) {
	var agentID int32
	fail := func(err error) (int32, error) {
		return 0, err
	}
	tx, err := d.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return fail(fmt.Errorf("%w, error in starting transaction", err))
	}
	defer tx.Rollback()

	getAvailableAgent := fmt.Sprintf("select id from delivery.agents where order_id is null and case when reserved_at is null then true else reserved_at < (now() - interval '%d' second) end limit 1", maxReserveTimeSec)
	row := tx.QueryRowContext(ctx, getAvailableAgent)
	if err := row.Scan(&agentID); err != nil {
		if err == sql.ErrNoRows {
			return fail(status.Error(codes.NotFound, "no agent found"))
		}
		return fail(fmt.Errorf("%w, error in getting the agent", err))
	}
	fmt.Printf("reserving agent id %d\n", agentID)
	if _, err := tx.ExecContext(ctx, `update delivery.agents set reserved_at = now() where id = $1`, agentID); err != nil {
		return fail(fmt.Errorf("%w, error in reserving the delivery agent", err))
	}

	if err := tx.Commit(); err != nil {
		return fail(fmt.Errorf("%w, failed to commit the transaction", err))
	}

	return agentID, nil
}

func (d *DB) BookTheAgent(ctx context.Context, agentId, orderId int32) error {
	var oid sql.NullInt32
	row := d.dbConn.QueryRowContext(ctx, `select order_id from delivery.agents where id = $1`, agentId)
	if err := row.Scan(&oid); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no such delivery agent found")
		}
		return fmt.Errorf("%w; something went wrong", err)
	}
	if oid.Valid && oid.Int32 != orderId {
		return fmt.Errorf("delivery agent is already assigned to another order")
	}

	if _, err := d.dbConn.ExecContext(ctx, `update delivery.agents set order_id = $1 where id = $2`, orderId, agentId); err != nil {
		return fmt.Errorf("%w; error in updating the order id to the delivery agent", err)
	}

	return nil
}
