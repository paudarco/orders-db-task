package database

import (
	"database/sql"

	"github.com/paudarco/orders-db-task/internal/models"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err = database.createTablesIfNotExist(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTablesIfNotExist() error {
	_, err := d.db.Exec(`
    CREATE TABLE IF NOT EXISTS orders (
        order_uid VARCHAR(255) PRIMARY KEY,
        track_number VARCHAR(255),
        entry VARCHAR(255),
        locale VARCHAR(10),
        internal_signature VARCHAR(255),
        customer_id VARCHAR(255),
        delivery_service VARCHAR(255),
        shardkey VARCHAR(255),
        sm_id INTEGER,
        date_created TIMESTAMP,
        oof_shard VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS deliveries (
        order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid),
        name VARCHAR(255),
        phone VARCHAR(20),
        zip VARCHAR(20),
        city VARCHAR(255),
        address VARCHAR(255),
        region VARCHAR(255),
        email VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS payments (
        transaction VARCHAR(255) PRIMARY KEY,
        order_uid VARCHAR(255) REFERENCES orders(order_uid),
        request_id VARCHAR(255),
        currency VARCHAR(10),
        provider VARCHAR(255),
        amount INTEGER,
        payment_dt BIGINT,
        bank VARCHAR(255),
        delivery_cost INTEGER,
        goods_total INTEGER,
        custom_fee INTEGER
    );

    CREATE TABLE IF NOT EXISTS items (
        chrt_id INTEGER,
        order_uid VARCHAR(255) REFERENCES orders(order_uid),
        track_number VARCHAR(255),
        price INTEGER,
        rid VARCHAR(255),
        name VARCHAR(255),
        sale INTEGER,
        size VARCHAR(10),
        total_price INTEGER,
        nm_id INTEGER,
        brand VARCHAR(255),
        status INTEGER
    );
    `)
	return err
}

func (d *Database) SaveOrder(order *models.Order) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	// Insert delivery
	_, err = tx.Exec(`
        INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	// Insert payment
	_, err = tx.Exec(`
        INSERT INTO payments (transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `, order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}

	// Insert items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `, item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) GetAllOrders() ([]*models.Order, error) {
	rows, err := d.db.Query(`
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
               d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
               p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        JOIN deliveries d ON o.order_uid = d.order_uid
        JOIN payments p ON o.order_uid = p.order_uid
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*models.Order, 0)
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			return nil, err
		}

		// Fetch items for this order
		itemRows, err := d.db.Query("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", order.OrderUID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item models.Item
			err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				return nil, err
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, &order)
	}

	return orders, nil
}
