package recipe

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	DB *pgx.Conn
}

func (r *Repo) List(ctx context.Context, hid int) ([]Recipe, error) {
	sql := `
        SELECT id, title, img_url, link, household_id
        FROM recipes
        WHERE household_id = $1;
    `

	rows, err := r.DB.Query(ctx, sql, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe

	for rows.Next() {
		var r Recipe
		err := rows.Scan(
			&r.Id,
			&r.Title,
			&r.ImgURL,
			&r.Link,
			&r.HouseholdID,
		)

		if err != nil {
			return nil, err
		}

		recipes = append(recipes, r)
	}

	return recipes, rows.Err()
}

func (r *Repo) AddRecipe(ctx context.Context, hid int, recipe Recipe) (int, error) {
	sql := `
	INSERT INTO recipes 
	(title, img_url, link, household_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`

	var id int
	err := r.DB.QueryRow(context.Background(), sql, recipe.Title, recipe.ImgURL, recipe.Link, recipe.HouseholdID).Scan(&id)

	return id, err
}
