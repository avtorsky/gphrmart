CREATE TABLE user_account (
	id SERIAL PRIMARY KEY,
	username VARCHAR UNIQUE NOT NULL,
  hashed_password VARCHAR NOT NULL,
  salt VARCHAR NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE order_status (
  id SERIAL PRIMARY KEY,
  status VARCHAR NOT NULL
);
INSERT INTO order_status(status) VALUES
('NEW'), ('REGISTERED'), ('PROCESSING'), ('INVALID'), ('PROCESSED');

CREATE TABLE user_order (
  id VARCHAR PRIMARY KEY,
  user_id INTEGER NOT NULL,
  status_id INTEGER DEFAULT 1 NOT NULL,
  uploaded_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES user_account(id),
  CONSTRAINT fk_status_id FOREIGN KEY (status_id) REFERENCES order_status(id)
);

CREATE TABLE transaction_type (
  id SERIAL PRIMARY KEY,
  type VARCHAR NOT NULL
);
INSERT INTO transaction_type(type) VALUES
('ACCRUAL'), ('WITHDRAWAL');

CREATE TABLE transaction(
  id SERIAL PRIMARY KEY,
  order_id VARCHAR NOT NULL,
  sum DOUBLE PRECISION NOT NULL,
  transaction_type_id INTEGER NOT NULL,
  processed_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT fk_transaction_type_id FOREIGN KEY (transaction_type_id) REFERENCES transaction_type(id),
  CONSTRAINT fk_order_id FOREIGN KEY (order_id) REFERENCES user_order(id)
);

CREATE TABLE accrual_queue(
  id SERIAL PRIMARY KEY,
  order_id VARCHAR NOT NULL,
  status_id INTEGER DEFAULT 1 NOT NULL,
  is_locked BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT fk_status_id FOREIGN KEY (status_id) REFERENCES order_status(id)
);