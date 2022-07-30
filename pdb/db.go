package pdb

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
)

var db *pgx.Conn

// connect to database using connection string as parameter.
func connectDB(c string) *pgx.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	conn, err := pgx.Connect(ctx, c)
	if err != nil {
		log.Fatal("cannot connect to database")
	}
	// ping to recheck if the connection is ready.
	err = conn.Ping(ctx)
	if err != nil {
		log.Fatal("database connection ping fail")
	}
	// later assign to global variable
	return conn
}

// check and create table for products if not exist
// receive conntext from outside function.
// inject db object from outside function in case need multiple table for difference db.
// exit program if error.
func migrateProduct(ctx context.Context) {
	// price 100 = 1 thb
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS products(
		id uuid PRIMARY KEY,
		name varchar(50) NOT NULL,
		price integer default 20000
	)`)
	if err != nil {
		log.Fatal("error, create product table. " + err.Error())
	}
}

func migrateProductTransactions(ctx context.Context) {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS product_transactions(
			id uuid PRIMARY KEY,
			product_id uuid REFERENCE products (id) NOT NULL,
			customer_id uuid REFERENCE customer_info (id) NOT NULL,
			action varchar(10) DEFAULT 'out',
			out_amount integer NOT NULL,
			created_on timestamp NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("error, create productTransactions table. " + err.Error())
	}
}

func migrateCustomerInfo(ctx context.Context) {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS customer_info(
		id uuid PRIMARY KEY,
		first_name varchar(100) NOT NULL,
		last_name varchar(100) NOT NULL,
		email varchar(155) NOT NULL,
		phone varchar(12)
	)`)
	if err != nil {
		log.Fatal("error , create customerInfo table. " + err.Error())
	}
}
