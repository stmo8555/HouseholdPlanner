package product

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	DB *pgx.Conn
}

func (r *Repo) Get(ctx context.Context, id int) (Product, error) {
	rows, err := r.DB.Query(
		ctx,
		`
		SELECT id, name, brand, store, category
		FROM products
		WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return Product{}, err
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])

	return p, err
}

func (r *Repo) GetID(ctx context.Context, p Product) (int, error) {
	var id int

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT id
		FROM products
		WHERE name=$1 AND brand=$2 AND store=$3;
		`,
		p.Name,
		p.Brand,
		p.Store,
	).Scan(&id)

	return id, err
}

func (r *Repo) Add(ctx context.Context, p Product) (int, error) {
	var id int

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO products (name, brand, store, category)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`,
		p.Name,
		p.Brand,
		p.Store,
		p.Category,
	).Scan(&id)

	return id, err
}
