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
			&r.Img_url,
			&r.Link,
			&r.Household_id,
		)

		if err != nil {
			return nil, err
		}

		recipes = append(recipes, r)
	}

	return recipes, rows.Err()
}

func (r *Repo) Add(ctx context.Context, hid int, recipe Recipe) error {
	sql := `INSERT INTO recipes 
	(title, img_url, link, household_id)
	VALUES ($1, $2, $3, $4);`

	_, err := r.DB.Exec(context.Background(), sql, recipe.Title, recipe.Img_url, recipe.Link, recipe.Household_id)

	return err
}
