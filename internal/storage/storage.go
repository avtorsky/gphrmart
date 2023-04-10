package storage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/avtorsky/gphrmart/internal/auth"
	"github.com/avtorsky/gphrmart/internal/config"
	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage/queue"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var OrderStatusMap = map[string]int{
	"NEW":        1,
	"REGISTERED": 2,
	"PROCESSING": 3,
	"INVALID":    4,
	"PROCESSED":  5,
}

type Storager interface {
	AddUser(ctx context.Context, username, password string) (*models.User, error)
	FindUser(ctx context.Context, username, password string) (*models.User, error)
	FindOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	AddOrder(ctx context.Context, orderID string, userID int) error
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	AddAccrualRecord(ctx context.Context, orderID string, sum float64) error
	FindOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)
	GetBalance(ctx context.Context, userID int) (*models.Balance, error)
	AddWithdrawalRecord(ctx context.Context, orderID string, sum float64, userID int) error
	GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	Queue() queue.OrderStatusNotifier
	Close()
}

type StorageService struct {
	conn *pgxpool.Pool
}

func NewStorageService(ctx context.Context, cfg *config.Config) (*StorageService, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	return &StorageService{conn: conn}, nil
}

func (db *StorageService) AddUser(ctx context.Context, username, password string) (*models.User, error) {
	log.Printf("Adding user %v...", username)
	salt, err := auth.GenerateSalt()
	if err != nil {
		return nil, err
	}
	hash := auth.GenerateHash(password, salt)
	var userID int
	err = db.conn.QueryRow(ctx, addUserQuery, username, hash, salt).Scan(&userID)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == pgerrcode.UniqueViolation {
				return nil, fmt.Errorf("%w: %v", ErrUserAlreadyExists, username)
			}
		}
		return nil, err
	}
	return &models.User{ID: userID, Username: username, HashedPassword: hash, Salt: salt}, nil
}

func (db *StorageService) FindUser(ctx context.Context, username, password string) (*models.User, error) {
	var hash, salt string
	var id int
	err := db.conn.QueryRow(ctx, getUserQuery, username).Scan(&id, &hash, &salt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", ErrUserNotFound, username)
		}
		return nil, err
	}
	return &models.User{ID: id, Username: username, HashedPassword: hash, Salt: salt}, nil
}

func (db *StorageService) FindOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	order := models.Order{}
	err := db.conn.QueryRow(ctx, getOrderByIDQuery, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.StatusID,
		&order.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, err
}

func (db *StorageService) AddOrder(ctx context.Context, orderID string, userID int) error {
	log.Printf("Adding order orderID=%v userID=%v...", orderID, userID)
	order, err := db.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := db.conn.Exec(ctx, addOrderQuery, orderID, userID)
			return err
		}
		return err
	}
	if order.UserID == userID {
		return fmt.Errorf("%w: orderID=%v userID=%v", ErrOrderAlreadyAddedByThisUser, orderID, userID)
	} else {
		return fmt.Errorf("%w: orderID=%v userID=%v", ErrOrderAlreadyAddedByOtherUser, orderID, userID)
	}
}

func (db *StorageService) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	_, err := db.conn.Exec(ctx, updateOrderStatusQuery, status, orderID)
	return err
}

func (db *StorageService) AddAccrualRecord(ctx context.Context, orderID string, sum float64) error {
	log.Printf("Adding accrual record orderID=%v sum=%v...", orderID, sum)
	_, err := db.conn.Exec(ctx, addTransactionQuery, orderID, sum, "ACCRUAL")
	return err
}

func (db *StorageService) FindOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	orders := make([]models.Order, 0)
	rows, err := db.conn.Query(ctx, getOrdersByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		order := models.Order{}
		if err := rows.Scan(&order.ID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrEmptyResult
	}
	return orders, nil
}

func (db *StorageService) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	balance := models.Balance{}
	err := db.conn.QueryRow(ctx, getBalanceQuery, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}
	return &balance, err
}

func (db *StorageService) AddWithdrawalRecord(
	ctx context.Context,
	orderID string,
	sum float64,
	userID int,
) error {
	log.Printf("Adding withdrawal record orderID=%v sum=%v...", orderID, sum)

	err := db.AddOrder(ctx, orderID, userID)
	if err != nil {
		if !errors.Is(err, ErrOrderAlreadyAddedByOtherUser) &&
			!errors.Is(err, ErrOrderAlreadyAddedByThisUser) {
			return err
		}
	}
	_, err = db.conn.Exec(ctx, addTransactionQuery, orderID, sum, "WITHDRAWAL")
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf("%w: %v", ErrOrderNotFound, orderID)
			}
		}
	}
	return err
}

func (db *StorageService) GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	withdrawals := make([]models.Withdrawal, 0)
	rows, err := db.conn.Query(ctx, getWithdrawalsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		withdrawal := models.Withdrawal{}
		if err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(withdrawals) == 0 {
		return nil, ErrEmptyResult
	}
	return withdrawals, nil
}

func (db *StorageService) Queue() queue.OrderStatusNotifier {
	return queue.NewOrderStatusQueue(db.conn)
}

func (db *StorageService) Close() {
	log.Println("Disconnecting storage...")
	db.conn.Close()
}
