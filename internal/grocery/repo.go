package grocery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
)

type Repo struct {
	db *pgxpool.Pool
}

func CreateRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}

	return &Repo{
		db: db,
	}
}

func (r *Repo) getTopProducts(ctx context.Context, householdID int) ([]product.Product, error) {
	sql := `
		SELECT p.id, p.name, p.brand, p.category
		FROM groceries_history gh
		INNER JOIN products p ON gh.product_id=p.id
		WHERE household_id = $1
		ORDER BY times_added DESC
		LIMIT 10;
	`

	rows, err := r.db.Query(ctx, sql, householdID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[product.Product])
}

func AddToHistory(tx pgx.Tx, ctx context.Context, grocerie Grocery) error {
	sql := `
	INSERT INTO groceries_history (household_id, product_id)
	VALUES ($1, $2)
	ON CONFLICT (household_id, product_id)
	DO UPDATE SET times_added = groceries_history.times_added + 1;
	`

	_, err := tx.Exec(ctx, sql, grocerie.HouseholdID, grocerie.Ingredient.ProductID)
	return err
}

func (r *Repo) AddGroceries(ctx context.Context, groceries []Grocery) error {

	sql := `
	INSERT INTO groceries 
	(product_id, amount, household_id)
	VALUES ($1, $2, $3)
	`

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, grocery := range groceries {
		_, err = tx.Exec(ctx, sql, grocery.Ingredient.ProductID, grocery.Ingredient.Amount, grocery.HouseholdID)
		if err != nil {
			return err
		}

		err = AddToHistory(tx, ctx, grocery)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func (r *Repo) List(ctx context.Context, sortBy, order string, householdID int) ([]Grocery, error) {
	sql := fmt.Sprintf(`
	SELECT 
		g.id,
		g.product_id,
		p.id,
		p.name,
		p.brand,
		p.category,
		g.amount,
		g.household_id,
		g.picked
	FROM groceries g
	INNER JOIN products p ON g.product_id = p.id
	WHERE g.household_id = $1
	ORDER BY %s %s
`, sortBy, order)

	rows, err := r.db.Query(ctx, sql, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groceries []Grocery

	for rows.Next() {
		var g Grocery

		err := rows.Scan(
			&g.Id,
			&g.Ingredient.ProductID,
			&g.Ingredient.Product.Id,
			&g.Ingredient.Product.Name,
			&g.Ingredient.Product.Brand,
			&g.Ingredient.Product.Category,
			&g.Ingredient.Amount,
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

func (r *Repo) TogglePicked(ctx context.Context, id, householdID int) error {
	sql := `
	UPDATE groceries
	SET picked = NOT picked 
	WHERE id=$1 AND household_id=$2;
	`
	_, err := r.db.Exec(ctx, sql, id, householdID)

	return err
}

func (r *Repo) Delete(ctx context.Context, groceryID, householdId int) error {
	sql := `
	DELETE FROM groceries
	WHERE id = $1 AND household_id = $2;
	`
	_, err := r.db.Exec(ctx, sql, groceryID, householdId)

	return err
}

func (r *Repo) DeletePicked(ctx context.Context, householdId int) error {
	sql := `
	DELETE FROM groceries
	WHERE household_id = $1 AND picked IS TRUE;
	`

	_, err := r.db.Exec(ctx, sql, householdId)
	return err
}

func (r *Repo) Edit(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	sql := `
	UPDATE groceries
    SET product_id=$1, amount=$2
    WHERE id=$3 AND household_id=$4;
	`
	_, err := r.db.Exec(ctx, sql, ing.ProductID, ing.Amount, groceryID, householdID)

	return err
}
