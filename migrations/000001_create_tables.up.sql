CREATE TABLE users
(
    id       serial PRIMARY KEY,
    login    varchar NOT NULL,
    password varchar NOT NULL,
    UNIQUE (login)
);

CREATE TABLE orders
(
    number      varchar PRIMARY KEY,
    user_id     integer     NOT NULL,
    status      varchar(16) NOT NULL,
    accrual     integer     NULL,
    uploaded_at timestamp DEFAULT timezone('utc', now()),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE transactions
(
    id           integer,
    order_number varchar,
    user_id      integer NOT NULL,
    sum          integer   NOT NULL,
    balance      integer   NOT NULL,
    withdrawals  integer   NOT NULL,
    processed_at timestamp DEFAULT timezone('utc', now()),
    FOREIGN KEY (user_id) REFERENCES users (id),
    UNIQUE (order_number),
    CONSTRAINT id_user_id PRIMARY KEY (user_id, id)
);

CREATE INDEX processing_order_index ON orders (status) WHERE status != 'PROCESSED' AND status != 'INVALID';
