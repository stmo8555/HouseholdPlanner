package recipe

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *Repo) List(ctx context.Context, hid int) ([]Recipe, map[int][]RecipeIngredient, error) {
	sql := `
	SELECT
	r.id,
	r.title,
	r.img_url,
	r.link,
	r.household_id,
	ri.id,
	ri.amount,
	p.id,
	p.name,
	p.brand,
	p.category
	FROM recipes r
	INNER JOIN recipe_ingredient ri ON ri.recipe_id = r.id
	INNER JOIN products p ON p.id = ri.product_id
	WHERE r.household_id = $1
	ORDER BY r.title;
	`

	rows, err := r.db.Query(ctx, sql, hid)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	recipeIngredients := make(map[int][]RecipeIngredient, 10)

	for rows.Next() {
		var r Recipe
		var ri RecipeIngredient
		err := rows.Scan(
			&r.Id,
			&r.Title,
			&r.ImgURL,
			&r.Link,
			&r.HouseholdID,
			&ri.Id,
			&ri.Ingredient.Amount,
			&ri.Ingredient.Product.Id,
			&ri.Ingredient.Product.Name,
			&ri.Ingredient.Product.Brand,
			&ri.Ingredient.Product.Category,
		)

		if err != nil {
			return nil, nil, err
		}

		if _, ok := recipeIngredients[r.Id]; !ok {
			recipes = append(recipes, r)
		}

		recipeIngredients[r.Id] = append(recipeIngredients[r.Id], ri)
	}

	return recipes, recipeIngredients, rows.Err()
}

func (r *Repo) Ingredients(ctx context.Context, recipeID, hid int) ([]RecipeIngredient, error) {
	sql := `
	SELECT
	ri.id,
	ri.amount,
	ri.product_id,
	p.id,
	p.name,
	p.brand,
	p.category
	FROM recipes r
	INNER JOIN recipe_ingredient ri ON ri.recipe_id=r.id
	INNER JOIN products p ON p.id=ri.product_id
	WHERE r.id=$1 and r.household_id=$2
	ORDER BY r.title;
	`
	rows, err := r.db.Query(ctx, sql, recipeID, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipeIngredients []RecipeIngredient

	for rows.Next() {
		var ri RecipeIngredient
		err := rows.Scan(
			&ri.Id,
			&ri.Ingredient.Amount,
			&ri.Ingredient.ProductID,
			&ri.Ingredient.Product.Id,
			&ri.Ingredient.Product.Name,
			&ri.Ingredient.Product.Brand,
			&ri.Ingredient.Product.Category,
		)

		if err != nil {
			return nil, err
		}

		recipeIngredients = append(recipeIngredients, ri)
	}

	return recipeIngredients, rows.Err()
}

func (r *Repo) AddRecipe(ctx context.Context, hid int, recipe Recipe) (int, error) {
	sql := `
	INSERT INTO recipes 
	(title, img_url, link, household_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`

	var id int
	err := r.db.QueryRow(context.Background(), sql, recipe.Title, recipe.ImgURL, recipe.Link, recipe.HouseholdID).Scan(&id)

	return id, err
}

func (r *Repo) AddIngredients(ctx context.Context, recipeIngredients []RecipeIngredient) error {
	sql := `
	INSERT INTO recipe_ingredient 
	(recipe_id, product_id, amount)
	VALUES ($1, $2, $3);
	`

	tx, err := r.db.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)

	for _, rec := range recipeIngredients {
		_, err := tx.Exec(ctx, sql, rec.RecipeID, rec.Ingredient.ProductID, rec.Ingredient.Amount)
		if err != nil {
			panic(err)
		}
	}
	return tx.Commit(ctx)
}
