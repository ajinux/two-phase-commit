-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA store;
CREATE TABLE store.food
(
    id   serial primary key,
    name varchar(100) not null
);
CREATE TABLE store.packet
(
    id          serial primary key,
    food_id     int references store.food (id),
    reserved_at timestamp,
    order_id    int
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA store CASCADE;
-- +goose StatementEnd
