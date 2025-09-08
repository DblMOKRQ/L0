package repository

import (
	"L0/models"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	orderQuery = `
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
                          customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            date_created = EXCLUDED.date_created,
            oof_shard = EXCLUDED.oof_shard
    `
	deliveryQuery = `
        INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email
    `
	paymentQuery = `
        INSERT INTO payments (order_uid, transaction, request_id, currency, provider, 
                             amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            transaction = EXCLUDED.transaction,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee
    `
	itemsQuery = `
        INSERT INTO items (
            order_uid, chrt_id, track_number, price, rid, name,
            sale, size, total_price, nm_id, brand, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (order_uid, rid) DO UPDATE SET
            chrt_id = EXCLUDED.chrt_id,
            track_number = EXCLUDED.track_number,
            price = EXCLUDED.price,
            name = EXCLUDED.name,
            sale = EXCLUDED.sale,
            size = EXCLUDED.size,
            total_price = EXCLUDED.total_price,
            nm_id = EXCLUDED.nm_id,
            brand = EXCLUDED.brand,
            status = EXCLUDED.status
    `
	orderQueryGet = `
        SELECT order_uid, track_number, entry, locale, internal_signature, 
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders 
        WHERE order_uid = $1
    `
	deliveryQueryGet = `
        SELECT name, phone, zip, city, address, region, email
        FROM deliveries 
        WHERE order_uid = $1
    `
	paymentQueryGet = `
        SELECT transaction, request_id, currency, provider, amount, 
               payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payments 
        WHERE order_uid = $1
    `
	itemsQueryGet = `
        SELECT chrt_id, track_number, price, rid, name, sale, size, 
               total_price, nm_id, brand, status
        FROM items 
        WHERE order_uid = $1
        ORDER BY id
    `
	recentGetQuery = `
		SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT $1
`
)

type Repository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func (s *Storage) NewRepository() *Repository {
	return &Repository{db: s.db, log: s.log.Named("Repository")}
}

func (r *Repository) SaveOrder(ctx context.Context, order *models.Order) error {
	r.log.Debug("Saving Order", zap.Any("order", order))
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		r.log.Error("Error begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Сохраняем основной заказ
	_, err = tx.Exec(ctx, orderQuery,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		r.log.Error("Error saving order", zap.Error(err))
		return fmt.Errorf("failed to save order: %w", err)
	}

	// Сохраняем доставку

	_, err = tx.Exec(ctx, deliveryQuery,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		r.log.Error("Error saving delivery order", zap.Error(err))
		return fmt.Errorf("failed to save delivery: %w", err)
	}

	// Сохраняем платеж

	_, err = tx.Exec(ctx, paymentQuery,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		r.log.Error("Error saving payment", zap.Error(err))
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// Сохраняем товары
	//batch := &pgx.Batch{}
	for _, item := range order.Items {
		_, err := tx.Exec(ctx, itemsQuery,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			r.log.Error("Error saving item", zap.Error(err))
			return fmt.Errorf("failed to save item: %w", err)
		}
	}
	r.log.Debug("Saved Order", zap.Any("order", order))
	return tx.Commit(ctx)
}

func (r *Repository) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	// Получаем основной заказ

	var order models.Order
	err := r.db.QueryRow(ctx, orderQueryGet, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn("Order not found", zap.String("order_uid", orderUID))
			return nil, models.OrderNotFoundError
		}
		r.log.Error("Error getting order", zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Получаем доставку

	err = r.db.QueryRow(ctx, deliveryQueryGet, orderUID).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		r.log.Error("Error getting delivery order", zap.Error(err))
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	// Получаем платеж

	err = r.db.QueryRow(ctx, paymentQueryGet, orderUID).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDt,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		r.log.Error("Error getting payment", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Получаем товары
	rows, err := r.db.Query(ctx, itemsQueryGet, orderUID)
	if err != nil {
		r.log.Error("Error getting items", zap.Error(err))
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	order.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			r.log.Error("Error getting item", zap.Error(err))
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating items", zap.Error(err))
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return &order, nil
}

func (r *Repository) GetRecentOrders(ctx context.Context, limit int) ([]*models.Order, error) {
	r.log.Debug("Getting recent orders ", zap.Int("limit", limit))
	var orderUIDs []string
	rows, err := r.db.Query(ctx, recentGetQuery, limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error("Error getting recent orders", zap.Error(err))
			return nil, fmt.Errorf("error getting recent orders: %w", err)
		}
		r.log.Error("Error getting recent orders", zap.Error(err))
		return nil, fmt.Errorf("error getting recent orders: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan order UID: %w", err)
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	var orders []*models.Order
	for _, uuid := range orderUIDs {
		order, err := r.GetOrderByUID(ctx, uuid)
		if err != nil {
			r.log.Error("Error getting order", zap.Error(err))
			return nil, fmt.Errorf("error getting order: %w", err)
		}
		orders = append(orders, order)
	}
	r.log.Debug("Received recent orders ", zap.Int("orders", len(orders)))
	return orders, nil

}

func (r *Repository) Close() {
	r.log.Info("Closing database")
	r.db.Close()
}
