package queue

const aquireQuery = `
UPDATE accrual_queue SET is_locked = TRUE
WHERE order_id = (
	SELECT order_id
	FROM accrual_queue
	WHERE status_id IN (1, 2, 3) AND is_locked = FALSE
	ORDER BY updated_at
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
RETURNING order_id, status_id;
`

const updateAndReleaseQuery = `
UPDATE accrual_queue
SET is_locked = FALSE, status_id = $1, updated_at = CURRENT_TIMESTAMP
WHERE order_id = $2;
`

const addQuery = `
INSERT INTO accrual_queue(order_id, status_id) VALUES ($1, DEFAULT);
`

const deleteQuery = `
DELETE FROM accrual_queue WHERE order_id = $1;
`
