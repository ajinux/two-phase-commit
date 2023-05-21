-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA delivery;
CREATE TABLE delivery.agents
(
    id          serial PRIMARY KEY,
    reserved_at timestamp,
    order_id    int
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA delivery CASCADE;
-- +goose StatementEnd
