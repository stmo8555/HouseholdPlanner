package grocery

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	DB *pgx.Conn
}

func (r *Repo) getTopProducts(ctx context.Context, householdID int) ([]string, error) {
	sql := `
		SELECT product
		FROM groceries_history
		WHERE household_id = $1
		ORDER BY times_added DESC
		LIMIT 10;
	`

	rows, err := r.DB.Query(ctx, sql, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []string

	for rows.Next() {
		var p string
		rows.Scan(&p)
		products = append(products, p)
	}

	return products, nil
}

func AddToHistory(conn *pgx.Conn, ctx context.Context, grocerie Grocery) error {
	sql := `INSERT INTO groceries_history (household_id, product)
	VALUES ($1, $2)
	ON CONFLICT (household_id, product)
	DO UPDATE SET times_added = groceries_history.times_added + 1;`

	_, err := conn.Exec(ctx, sql, grocerie.HouseholdID, grocerie.Product)
	return err
}

func (r *Repo) AddGroceries(ctx context.Context, groceries []Grocery) error {

	sql := `INSERT INTO groceries 
	(product, brand, amount, store, picked, household_id)
	VALUES ($1, $2, $3, $4, FALSE, $5)`

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, grocery := range groceries {
		_, err = tx.Exec(ctx, sql, grocery.Product, grocery.Brand, grocery.Amount, grocery.Store, grocery.HouseholdID)
		if err != nil {
			return err
		}

		err = AddToHistory(tx.Conn(), ctx, grocery)
		
		if err != nil {
			return err
		}
	}

	return  tx.Commit(ctx)
}
func (r *Repo) List(ctx context.Context, householdID int) ([]Grocery, error) {
	sql := `
        SELECT id, product, brand, store, amount, household_id, picked 
        FROM groceries
        WHERE household_id = $1
        ORDER BY product;`

	rows, err := r.DB.Query(ctx, sql, householdID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var groceries []Grocery

	for rows.Next() {
		var g Grocery
		err := rows.Scan(
			&g.Id,
			&g.Product,
			&g.Brand,
			&g.Store,
			&g.Amount,
			&g.HouseholdID,
			&g.Picked,
		)

		if err != nil {
			return nil, err
		}

		groceries = append(groceries, g)
	}

	return groceries, rows.Err()
}

func (r *Repo) AmountOfUnpickedGroceries(ctx context.Context, hid int) (int, error) {
	sql := `
        SELECT COUNT(*)
        FROM groceries
        WHERE household_id = $1 AND NOT picked;
    `

	var count int

	err := r.DB.QueryRow(ctx, sql, hid).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repo) TogglePicked(ctx context.Context,id, householdID int) error {
	sql := `UPDATE groceries SET picked = NOT picked WHERE id=$1 AND household_id=$2;`
	_, err := r.DB.Exec(ctx, sql, id, householdID)

	return err
} 

func (r *Repo) DeletePicked(ctx context.Context, householdId int) error {
	sql := `DELETE FROM groceries
			WHERE household_id = $1 AND picked IS TRUE;`

	_, err := r.DB.Exec(ctx, sql, householdId)
	return err
}

func (r *Repo) Edit(ctx context.Context, groceries []Grocery) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	for _, g := range groceries {
		sql := `UPDATE groceries
               SET product=$1, amount=$2, brand=$3, store=$4
               WHERE id=$5`
		_, err = tx.Exec(context.Background(), sql, g.Product, g.Amount, g.Brand, g.Store, g.Id)
	}

	return tx.Commit(context.Background())
}
