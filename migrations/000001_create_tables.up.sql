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
    accrual     float       NULL,
    uploaded_at timestamp DEFAULT timezone('utc', now()),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE transactions
(
    id           integer,
    number       varchar,
    user_id      integer NOT NULL,
    sum          float   NOT NULL,
    balance      float   NOT NULL,
    withdrawals  float   NOT NULL,
    processed_at timestamp DEFAULT timezone('utc', now()),
    FOREIGN KEY (user_id) REFERENCES users (id),
    UNIQUE (number),
    CONSTRAINT id_user_id PRIMARY KEY (user_id, id)
);

CREATE INDEX processing_order_index ON orders(status) WHERE status != 'PROCESSED' AND status != 'INVALID';
