package storage

const addUserQuery = `
INSERT INTO user_account(username, hashed_password, salt) VALUES ($1, $2, $3) RETURNING id;
`

const getUserQuery = `
SELECT id, hashed_password, salt FROM user_account WHERE username = $1;
`

const addOrderQuery = `
INSERT INTO user_order(id, user_id, status_id) VALUES ($1, $2, DEFAULT);
`

const getOrderByIDQuery = `
SELECT id, user_id, status_id, uploaded_at FROM user_order WHERE id = $1;
`

const updateOrderStatusQuery = `
UPDATE user_order SET status_id = (SELECT id FROM order_status WHERE status = $1) WHERE id = $2;
`

const addTransactionQuery = `
INSERT INTO transaction(order_id, sum, transaction_type_id)
	SELECT $1, $2, id
	FROM transaction_type
	WHERE type = $3;
`

const getOrdersByUserIDQuery = `
WITH a AS (
	SELECT order_id, SUM(sum) as sum
	FROM transaction
	WHERE transaction_type_id = 1
	GROUP BY order_id
)
SELECT b.id AS "number", c.status, a.sum AS "accrual", b.uploaded_at
FROM user_order b
LEFT JOIN a ON b.id = a.order_id
JOIN order_status c ON c.id = b.status_id
WHERE b.user_id = $1
ORDER BY b.uploaded_at;
`

const getBalanceQuery = `
SELECT (
	COALESCE(SUM(CASE WHEN transaction_type_id = 1 THEN a.sum END), 0) -
	COALESCE(SUM(CASE WHEN transaction_type_id = 2 THEN a.sum END), 0)
) AS balance, (
	COALESCE(SUM(CASE WHEN transaction_type_id = 2 THEN a.sum END), 0)
) AS withdrawn
FROM transaction a
JOIN user_order b ON b.id = a.order_id
WHERE b.user_id = $1;
`

const getWithdrawalsQuery = `
SELECT a.order_id AS order, a.sum, a.processed_at
FROM transaction a
JOIN user_order b ON b.id = a.order_id
WHERE b.user_id = $1 AND a.transaction_type_id = 2
ORDER BY a.processed_at;
`
