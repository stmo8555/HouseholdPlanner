package product

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}
	return &Repo{
		db: db,
	}
}

func (r *Repo) Get(ctx context.Context, id int) (Product, error) {
	sql := `
		SELECT id, name, brand, category
		FROM products
		WHERE id = $1
		`
	rows, err := r.db.Query(
		ctx,
		sql,
		id,
	)
	if err != nil {
		return Product{}, err
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])

	return p, err
}

func (r *Repo) GetID(ctx context.Context, p Product) (int, error) {
	sql := `
		SELECT id
		FROM products
		WHERE name=$1 AND brand=$2;
		`
	var id int

	err := r.db.QueryRow(
		ctx,
		sql,
		p.Name,
		p.Brand,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return id, ErrNotFound
	}

	return id, err
}

func (r *Repo) Add(ctx context.Context, p Product) (int, error) {

	sql := `
		INSERT INTO products (name, brand, category)
		VALUES ($1, $2, $3)
		RETURNING id
		`
	var id int

	err := r.db.QueryRow(
		ctx,
		sql,
		p.Name,
		p.Brand,
		p.Category,
	).Scan(&id)

	return id, err
}
