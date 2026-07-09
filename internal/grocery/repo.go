package grocery

import (
	"context"
	"errors"
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

func (r *Repo) GroceryList(ctx context.Context, groceryListID, hid int) (GroceryList, error) {
	sql := `
	SELECT id, name, household_id 
	FROM grocery_lists
	WHERE id = $1 AND household_id = $2;
	`

	var g GroceryList
	row := r.db.QueryRow(ctx, sql, groceryListID, hid)
	err := row.Scan(
		&g.ID,
		&g.Name,
		&g.HouseholdID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return g, ErrNotFound
	}

	return g, err
}


func (r *Repo) GroceryListsStats(ctx context.Context, hid int) (map[int]GroceryListStats, error) {
	sql := `
	SELECT
		gl.id,
		COALESCE(hpc.category, p.category) AS category,
		COUNT(g.id) AS total
	FROM grocery_lists gl
	LEFT JOIN groceries g ON g.grocery_list_id = gl.id
	LEFT JOIN products p ON p.id = g.product_id
	LEFT JOIN household_product_category hpc
		ON hpc.household_id = gl.household_id AND hpc.product_id = g.product_id
	WHERE gl.household_id = $1
	GROUP BY gl.id, COALESCE(hpc.category, p.category)
	ORDER BY gl.id, total DESC, category;
	`

	rows, err := r.db.Query(ctx, sql, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[int]GroceryListStats)

	for rows.Next() {
		var listID int
		var category *string
		var total int

		if err := rows.Scan(&listID, &category, &total); err != nil {
			return nil, err
		}

		g := stats[listID]
		g.ListID = listID
		g.Total += total

		if category != nil && total > 0 {
			g.Categories = append(g.Categories, CategoryCount{Label: *category, Count: total})
		}

		stats[listID] = g
	}

	return stats, rows.Err()
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

func (r *Repo) UpdateGroceryList(ctx context.Context, newName string, groceryListID int, hid int) error {
	sql := `
	UPDATE grocery_lists 
	SET name = $1
	WHERE id = $2 AND household_id = $3;
	`
	_, err := r.db.Exec(ctx, sql, newName, groceryListID, hid)

	return err
}

func (r *Repo) DeleteGroceryList(ctx context.Context, groceryListID int, hid int) error {
	sql := `
	DELETE FROM grocery_lists
	WHERE id = $1 AND household_id = $2;
	`
	_, err := r.db.Exec(ctx, sql, groceryListID, hid)

	return err
}

func (r *Repo) TransferGroceries(ctx context.Context, groceryListTargetID int, groceryListID int, hid int) error {
	sql := `
    UPDATE groceries
    SET grocery_list_id = $1
    WHERE grocery_list_id = $2
      AND EXISTS (
        SELECT 1 FROM grocery_lists
        WHERE id = $2 AND household_id = $3
      )
      AND EXISTS (
        SELECT 1 FROM grocery_lists
        WHERE id = $1 AND household_id = $3
      );
    `
	res, err := r.db.Exec(ctx, sql, groceryListTargetID, groceryListID, hid)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("target list not found in household or source list is empty")
	}

	return nil
}

func (r *Repo) MoveGrocery(ctx context.Context, groceryID int, groceryListTargetID int, hid int) error {
	sql := `
    UPDATE groceries
    SET grocery_list_id = $1
    WHERE id = $2
      AND EXISTS (
        SELECT 1 FROM grocery_lists
        WHERE id = groceries.grocery_list_id AND household_id = $3
      )
      AND EXISTS (
        SELECT 1 FROM grocery_lists
        WHERE id = $1 AND household_id = $3
      );
    `
	res, err := r.db.Exec(ctx, sql, groceryListTargetID, groceryID, hid)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("grocery or target list not found in household")
	}

	return nil
}

func (r *Repo) CreateGroceries(ctx context.Context, groceries []Grocery, groceryListID, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var owned bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM grocery_lists WHERE id = $1 AND household_id = $2)`,
		groceryListID, hid).Scan(&owned)
	if err != nil {
		return err
	}
	if !owned {
		return ErrNotFound
	}

	sql := `
	INSERT INTO groceries
	(product_id, amount, grocery_list_id)
	VALUES ($1, $2, $3)
	`

	for _, grocery := range groceries {
		_, err = tx.Exec(ctx, sql, grocery.Ingredient.ProductID, grocery.Ingredient.Amount, grocery.ListID)
		if err != nil {
			return err
		}

		err = upsertHistory(tx, ctx, hid, grocery.Ingredient.ProductID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repo) Groceries(ctx context.Context, sortBy, order string, groceryListID, householdID int) ([]Grocery, error) {
	sql := fmt.Sprintf(`
	SELECT
		g.id,
		g.product_id,
		p.id,
		p.name,
		p.brand,
		COALESCE(hpc.category, p.category),
		g.amount,
		g.grocery_list_id,
		g.picked
	FROM groceries g
	INNER JOIN products p ON g.product_id = p.id
	INNER JOIN grocery_lists gl ON gl.id = g.grocery_list_id
	LEFT JOIN household_product_category hpc
		ON hpc.household_id = gl.household_id AND hpc.product_id = g.product_id
	WHERE g.grocery_list_id = $1 AND gl.household_id = $2
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
			&g.ListID,
			&g.Picked,
		)
		if err != nil {
			return nil, err
		}

		groceries = append(groceries, g)
	}

	return groceries, rows.Err()
}

func (r *Repo) Grocery(ctx context.Context, itemID int, hid int) (Grocery, error) {
	sql := `
	SELECT
		g.id,
		g.product_id,
		p.id,
		p.name,
		p.brand,
		COALESCE(hpc.category, p.category),
		g.amount,
		g.grocery_list_id,
		g.picked
	FROM groceries g
	INNER JOIN products p ON g.product_id = p.id
	INNER JOIN grocery_lists gl ON gl.id = g.grocery_list_id
	LEFT JOIN household_product_category hpc
		ON hpc.household_id = gl.household_id AND hpc.product_id = g.product_id
	WHERE g.id = $1 AND gl.household_id = $2
	`

	var g Grocery
	row := r.db.QueryRow(ctx, sql, itemID, hid)
	err := row.Scan(
		&g.ID,
		&g.Ingredient.ProductID,
		&g.Ingredient.Product.Id,
		&g.Ingredient.Product.Name,
		&g.Ingredient.Product.Brand,
		&g.Ingredient.Product.Category,
		&g.Ingredient.Amount,
		&g.ListID,
		&g.Picked,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return g, ErrNotFound
	}

	return g, err
}

func (r *Repo) UpdateGrocery(ctx context.Context, ing ingredient.Ingredient, groceryID, householdID int) error {
	sql := `
	UPDATE groceries
    SET product_id=$1, amount=$2
    WHERE id=$3
      AND EXISTS (
        SELECT 1 FROM grocery_lists gl
        WHERE gl.id = groceries.grocery_list_id AND gl.household_id = $4
      );
	`
	_, err := r.db.Exec(ctx, sql, ing.ProductID, ing.Amount, groceryID, householdID)

	return err
}

func (r *Repo) SetCategoryOverride(ctx context.Context, householdID, productID int, category string) error {
	sql := `
	INSERT INTO household_product_category (household_id, product_id, category)
	VALUES ($1, $2, $3)
	ON CONFLICT (household_id, product_id)
	DO UPDATE SET category = EXCLUDED.category;
	`
	_, err := r.db.Exec(ctx, sql, householdID, productID, category)

	return err
}

func (r *Repo) DeleteGrocery(ctx context.Context, groceryID, householdId int) error {
	sql := `
	DELETE FROM groceries
	WHERE id = $1
	  AND EXISTS (
	    SELECT 1 FROM grocery_lists gl
	    WHERE gl.id = groceries.grocery_list_id AND gl.household_id = $2
	  );
	`
	_, err := r.db.Exec(ctx, sql, groceryID, householdId)

	return err
}

func (r *Repo) DeletePicked(ctx context.Context, groceryListID, householdID int) error {
	sql := `
	DELETE FROM groceries
	WHERE grocery_list_id = $1 AND picked IS TRUE
	  AND EXISTS (
	    SELECT 1 FROM grocery_lists gl
	    WHERE gl.id = $1 AND gl.household_id = $2
	  );
	`

	_, err := r.db.Exec(ctx, sql, groceryListID, householdID)

	return err
}

func (r *Repo) TogglePicked(ctx context.Context, id, householdID int) error {
	sql := `
	UPDATE groceries
	SET picked = NOT picked
	WHERE id=$1
	  AND EXISTS (
	    SELECT 1 FROM grocery_lists gl
	    WHERE gl.id = groceries.grocery_list_id AND gl.household_id = $2
	  );
	`
	_, err := r.db.Exec(ctx, sql, id, householdID)

	return err
}

// SetPicked is the idempotent variant used by offline sync replay; a queued
// change applied twice (or against an item deleted meanwhile) is a no-op.
func (r *Repo) SetPicked(ctx context.Context, id, householdID int, picked bool) error {
	sql := `
	UPDATE groceries
	SET picked = $3
	WHERE id=$1
	  AND EXISTS (
	    SELECT 1 FROM grocery_lists gl
	    WHERE gl.id = groceries.grocery_list_id AND gl.household_id = $2
	  );
	`
	_, err := r.db.Exec(ctx, sql, id, householdID, picked)

	return err
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

func upsertHistory(tx pgx.Tx, ctx context.Context, hid, productID int) error {
	sql := `
	INSERT INTO groceries_history (household_id, product_id)
	VALUES ($1, $2)
	ON CONFLICT (household_id, product_id)
	DO UPDATE SET times_added = groceries_history.times_added + 1;
	`

	_, err := tx.Exec(ctx, sql, hid, productID)

	return err
}
