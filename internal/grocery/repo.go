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

func NewRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}

	return &Repo{
		db: db,
	}
}

func (r *Repo) TopProducts(ctx context.Context, householdID int) ([]product.Product, error) {
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

func upsertHistory(tx pgx.Tx, ctx context.Context, grocery Grocery) error {
	sql := `
	INSERT INTO groceries_history (household_id, product_id)
	VALUES ($1, $2)
	ON CONFLICT (household_id, product_id)
	DO UPDATE SET times_added = groceries_history.times_added + 1;
	`

	_, err := tx.Exec(ctx, sql, grocery.HouseholdID, grocery.Ingredient.ProductID)
	return err
}

func (r *Repo) CreateList(ctx context.Context, name string, hid int) error {
	sql := `
	INSERT INTO grocery_lists 
	(name, household_id)
	VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, sql, name, hid)

	return err
}

func (r *Repo) CreateGroceries(ctx context.Context, groceries []Grocery) error {

	sql := `
	INSERT INTO groceries 
	(product_id, amount, grocery_list_id, household_id)
	VALUES ($1, $2, $3, $4)
	`

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, grocery := range groceries {
		_, err = tx.Exec(ctx, sql, grocery.Ingredient.ProductID, grocery.Ingredient.Amount, grocery.GroceryListID, grocery.HouseholdID)
		if err != nil {
			return err
		}

		err = upsertHistory(tx, ctx, grocery)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func (r *Repo) GroceryListsStats(ctx context.Context, hid int) (map[int]GroceryListStats, error) {
	sql := `
	SELECT
    	gl.id,
    	COUNT(g.id) AS total,
    	COUNT(g.id) FILTER (WHERE g.picked) AS picked
	FROM grocery_lists gl
	LEFT JOIN groceries g ON g.grocery_list_id = gl.id
	WHERE gl.household_id = $1
	GROUP BY gl.id
	ORDER BY gl.id;
	`

	rows, err := r.db.Query(ctx, sql, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groceryListsStats := make(map[int]GroceryListStats)

	for rows.Next() {
		var g GroceryListStats

		err := rows.Scan(
			&g.ListID,
			&g.Total,
			&g.Picked,
		)
		if err != nil {
			return nil, err
		}

		groceryListsStats[g.ListID] = g
	}
	return groceryListsStats, nil
}
func (r *Repo) GroceryLists(ctx context.Context, hid int) ([]GroceryList, error) {
	sql := `
	SELECT id, name, household_id FROM grocery_lists
	WHERE household_id = $1;
	`

	rows, err := r.db.Query(ctx, sql, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groceryLists []GroceryList

	for rows.Next() {
		var g GroceryList

		err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.HouseholdID,
		)
		if err != nil {
			return nil, err
		}

		groceryLists = append(groceryLists, g)
	}

	return groceryLists, rows.Err()
}

func (r *Repo) Groceries(ctx context.Context, sortBy, order string, groceryListID, householdID int) ([]Grocery, error) {
	sql := fmt.Sprintf(`
	SELECT 
		g.id,
		g.product_id,
		p.id,
		p.name,
		p.brand,
		p.category,
		g.amount,
		g.grocery_list_id,
		g.household_id,
		g.picked
	FROM groceries g
	INNER JOIN products p ON g.product_id = p.id
	WHERE g.grocery_list_id = $1 AND g.household_id = $2
	ORDER BY %s %s
`, sortBy, order)

	rows, err := r.db.Query(ctx, sql, groceryListID, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groceries []Grocery

	for rows.Next() {
		var g Grocery

		err := rows.Scan(
			&g.ID,
			&g.Ingredient.ProductID,
			&g.Ingredient.Product.Id,
			&g.Ingredient.Product.Name,
			&g.Ingredient.Product.Brand,
			&g.Ingredient.Product.Category,
			&g.Ingredient.Amount,
			&g.GroceryListID,
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

func (r *Repo) DeleteGrocery(ctx context.Context, groceryID, householdId int) error {
	sql := `
	DELETE FROM groceries
	WHERE id = $1 AND household_id = $2;
	`
	_, err := r.db.Exec(ctx, sql, groceryID, householdId)

	return err
}

func (r *Repo) DeletePicked(ctx context.Context, groceryListID, householdID int) error {
	sql := `
	DELETE FROM groceries
	WHERE grocery_list_id = $1 AND household_id = $2 AND picked IS TRUE;
	`

	_, err := r.db.Exec(ctx, sql, groceryListID, householdID)
	return err
}

func (r *Repo) UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	sql := `
	UPDATE groceries
    SET product_id=$1, amount=$2
    WHERE id=$3 AND household_id=$4;
	`
	_, err := r.db.Exec(ctx, sql, ing.ProductID, ing.Amount, groceryID, householdID)

	return err
}
