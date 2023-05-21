-- +goose Up
-- +goose StatementBegin
-- Add 7 delivery agent who are free to take the order
INSERT INTO delivery.agents (reserved_at, order_id)
values (null, null),
       (null, null),
       (null, null),
       (null, null),
       (null, null),
       (null, null),
       (null, null);

INSERT INTO store.food (name)
values ('shawarma'),
       ('Dosa'),
       ('Burger'),
       ('pizza');
-- +goose StatementEnd
-- +goose StatementBegin
-- Add the number of packets available for the food
INSERT INTO store.packet (food_id)
values (1),
       (2),
       (1),
       (1),
       (3),
       (1),
       (1),
       (2),
       (4),
       (4),
       (2),
       (1),
       (3),
       (3),
       (1),
       (2),
       (4),
       (1);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- RESTART IDENTITY will restart the sequence from start
TRUNCATE TABLE delivery.agents RESTART IDENTITY;
TRUNCATE TABLE store.packet RESTART IDENTITY;
TRUNCATE TABLE store.food RESTART IDENTITY CASCADE;
-- +goose StatementEnd
