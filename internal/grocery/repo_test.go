package grocery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stmo8555/HouseholdPlanner/internal/ingredient"
)

// testPool is a connection to an ephemeral Postgres started in TestMain. These
// tests exercise the real SQL scoping (the household ownership guards), so they
// need an actual database rather than a mocked IRepo. If docker is unavailable
// the whole package's DB tests are skipped.
var (
	testPool   *pgxpool.Pool
	testUserID int
)

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("grocery: docker not found, skipping DB integration tests")
		os.Exit(0)
	}

	name := fmt.Sprintf("hp_grocery_test_%d", os.Getpid())
	_ = exec.Command("docker", "rm", "-f", name).Run()

	out, err := exec.Command("docker", "run", "-d", "--name", name,
		"-e", "POSTGRES_USER=test", "-e", "POSTGRES_PASSWORD=test", "-e", "POSTGRES_DB=test",
		"-p", "127.0.0.1::5432", "postgres:18").CombinedOutput()
	if err != nil {
		fmt.Printf("grocery: could not start postgres, skipping DB tests: %v\n%s\n", err, out)
		os.Exit(0)
	}

	port, err := mappedPort(name)
	if err != nil {
		teardownFatal(name, "resolve mapped port", err)
	}
	if err := waitReady(name); err != nil {
		teardownFatal(name, "wait for postgres", err)
	}
	if err := applySchema(name); err != nil {
		teardownFatal(name, "apply init.sql", err)
	}

	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf("postgres://test:test@127.0.0.1:%s/test?sslmode=disable", port))
	if err != nil {
		teardownFatal(name, "connect", err)
	}
	testPool = pool

	// init.sql creates no users; the real deployment seeds one from
	// zz-admin-user.sh, which needs env this harness does not provide.
	if err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, pwd) VALUES ('test-owner', 'x') RETURNING id`,
	).Scan(&testUserID); err != nil {
		teardownFatal(name, "seed user", err)
	}

	code := m.Run()

	pool.Close()
	_ = exec.Command("docker", "rm", "-f", name).Run()
	os.Exit(code)
}

func teardownFatal(name, what string, err error) {
	_ = exec.Command("docker", "rm", "-f", name).Run()
	fmt.Printf("grocery: %s failed: %v\n", what, err)
	os.Exit(1)
}

func mappedPort(name string) (string, error) {
	out, err := exec.Command("docker", "port", name, "5432/tcp").Output()
	if err != nil {
		return "", err
	}
	// e.g. "127.0.0.1:49153" (possibly followed by an IPv6 line)
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	i := strings.LastIndex(line, ":")
	if i < 0 {
		return "", fmt.Errorf("unexpected docker port output: %q", out)
	}
	return line[i+1:], nil
}

func waitReady(name string) error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		// Not pg_isready: it also answers the temporary server initdb runs
		// during bootstrap, before the real database exists. A query against
		// "test" only succeeds once the container is genuinely serving it.
		if exec.Command("docker", "exec", name,
			"psql", "-U", "test", "-d", "test", "-c", "SELECT 1").Run() == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("postgres not ready within timeout")
}

func applySchema(name string) error {
	f, err := os.Open("../../household-db/db-init/init.sql")
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("docker", "exec", "-i", name,
		"psql", "-U", "test", "-d", "test", "-v", "ON_ERROR_STOP=1", "-q")
	cmd.Stdin = f
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v\n%s", err, out)
	}
	return nil
}

var uniq int64

func next() int64 { return atomic.AddInt64(&uniq, 1) }

func insertHousehold(t *testing.T, label string) int {
	t.Helper()
	var id int
	n := next()
	if err := testPool.QueryRow(context.Background(),
		`INSERT INTO households (name, code, created_by) VALUES ($1, $2, $3) RETURNING id`,
		fmt.Sprintf("%s-%d", label, n), fmt.Sprintf("%06d", n), testUserID).Scan(&id); err != nil {
		t.Fatalf("insert household: %v", err)
	}
	return id
}

func insertList(t *testing.T, hid int) int {
	t.Helper()
	var id int
	if err := testPool.QueryRow(context.Background(),
		`INSERT INTO grocery_lists (name, household_id) VALUES ($1, $2) RETURNING id`,
		fmt.Sprintf("list-%d", next()), hid).Scan(&id); err != nil {
		t.Fatalf("insert list: %v", err)
	}
	return id
}

func insertProduct(t *testing.T) int {
	t.Helper()
	var id int
	n := next()
	if err := testPool.QueryRow(context.Background(),
		`INSERT INTO products (name, category) VALUES ($1, 'other') RETURNING id`,
		fmt.Sprintf("prod-%d", n)).Scan(&id); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	return id
}

func countOverrides(t *testing.T, hid, productID int) int {
	t.Helper()
	var n int
	if err := testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM household_product_category WHERE household_id = $1 AND product_id = $2`,
		hid, productID).Scan(&n); err != nil {
		t.Fatalf("count overrides: %v", err)
	}
	return n
}

func countItems(t *testing.T, listID int) int {
	t.Helper()
	var n int
	if err := testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM groceries WHERE grocery_list_id = $1`, listID).Scan(&n); err != nil {
		t.Fatalf("count items: %v", err)
	}
	return n
}

func milk(listID, productID int) []Grocery {
	return []Grocery{{
		Ingredient: ingredient.Ingredient{ProductID: productID, Amount: "1"},
		ListID:     listID,
	}}
}

// TestCreateGroceries_RejectsForeignList covers the up-front ownership check: a
// household must not be able to insert items into another household's list via a
// guessed list id, and a rejected insert must leave no rows behind.
func TestCreateGroceries_RejectsForeignList(t *testing.T) {
	repo := NewRepo(testPool)
	ctx := context.Background()

	hOwner := insertHousehold(t, "owner")
	hOther := insertHousehold(t, "other")
	list := insertList(t, hOwner)
	prod := insertProduct(t)

	// Foreign household: must be rejected with ErrNotFound and insert nothing.
	if err := repo.CreateGroceries(ctx, milk(list, prod), list, hOther); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign insert: got err %v, want ErrNotFound", err)
	}
	if n := countItems(t, list); n != 0 {
		t.Fatalf("foreign insert must roll back, found %d row(s) in list", n)
	}

	// Owning household: succeeds.
	if err := repo.CreateGroceries(ctx, milk(list, prod), list, hOwner); err != nil {
		t.Fatalf("owner insert: unexpected err %v", err)
	}
	if n := countItems(t, list); n != 1 {
		t.Fatalf("owner insert: want 1 row, got %d", n)
	}
}

// TestDeleteGrocery_ScopedToHousehold covers the inline EXISTS guard: a foreign
// household's delete is a no-op, the owner's delete works.
func TestDeleteGrocery_ScopedToHousehold(t *testing.T) {
	repo := NewRepo(testPool)
	ctx := context.Background()

	hOwner := insertHousehold(t, "owner")
	hOther := insertHousehold(t, "other")
	list := insertList(t, hOwner)
	prod := insertProduct(t)

	if err := repo.CreateGroceries(ctx, milk(list, prod), list, hOwner); err != nil {
		t.Fatalf("seed grocery: %v", err)
	}
	var itemID int
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM groceries WHERE grocery_list_id = $1 LIMIT 1`, list).Scan(&itemID); err != nil {
		t.Fatalf("fetch seeded item: %v", err)
	}

	// Foreign household: delete affects nothing.
	if err := repo.DeleteGrocery(ctx, itemID, hOther); err != nil {
		t.Fatalf("foreign delete: unexpected err %v", err)
	}
	if n := countItems(t, list); n != 1 {
		t.Fatalf("foreign delete must not remove the row, got %d", n)
	}

	// Owning household: delete works.
	if err := repo.DeleteGrocery(ctx, itemID, hOwner); err != nil {
		t.Fatalf("owner delete: unexpected err %v", err)
	}
	if n := countItems(t, list); n != 0 {
		t.Fatalf("owner delete: want 0 rows, got %d", n)
	}
}

// TestSetCategoryOverride_ClearedWhenBackToDefault covers the round trip: moving a
// product away from its default stores an override, moving it back removes the row
// rather than storing one that merely duplicates the default.
func TestSetCategoryOverride_ClearedWhenBackToDefault(t *testing.T) {
	repo := NewRepo(testPool)
	ctx := context.Background()

	hid := insertHousehold(t, "override")
	list := insertList(t, hid)
	prod := insertProduct(t) // seeded with category 'other'

	if err := repo.CreateGroceries(ctx, milk(list, prod), list, hid); err != nil {
		t.Fatalf("seed grocery: %v", err)
	}
	var itemID int
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM groceries WHERE grocery_list_id = $1 LIMIT 1`, list).Scan(&itemID); err != nil {
		t.Fatalf("fetch seeded item: %v", err)
	}

	// Away from the default: the override is stored and takes effect.
	if err := repo.SetCategoryOverride(ctx, hid, prod, "Dairy"); err != nil {
		t.Fatalf("set override: %v", err)
	}
	if n := countOverrides(t, hid, prod); n != 1 {
		t.Fatalf("want 1 override row, got %d", n)
	}
	g, err := repo.Grocery(ctx, itemID, hid)
	if err != nil {
		t.Fatalf("read grocery: %v", err)
	}
	if g.Ingredient.Product.Category != "Dairy" {
		t.Fatalf("want effective category Dairy, got %q", g.Ingredient.Product.Category)
	}

	// Back to the default: the row is removed, not rewritten.
	if err := repo.SetCategoryOverride(ctx, hid, prod, "other"); err != nil {
		t.Fatalf("reset override: %v", err)
	}
	if n := countOverrides(t, hid, prod); n != 0 {
		t.Fatalf("want the override removed, got %d rows", n)
	}
	g, err = repo.Grocery(ctx, itemID, hid)
	if err != nil {
		t.Fatalf("read grocery: %v", err)
	}
	if g.Ingredient.Product.Category != "other" {
		t.Fatalf("want effective category other, got %q", g.Ingredient.Product.Category)
	}
}

// TestSetCategoryOverride_NoOpWhenAlreadyDefault covers the case where no override
// exists and the default is re-selected: the delete must not fail on a missing row.
func TestSetCategoryOverride_NoOpWhenAlreadyDefault(t *testing.T) {
	repo := NewRepo(testPool)
	ctx := context.Background()

	hid := insertHousehold(t, "override-noop")
	prod := insertProduct(t)

	if err := repo.SetCategoryOverride(ctx, hid, prod, "other"); err != nil {
		t.Fatalf("set override to default: %v", err)
	}
	if n := countOverrides(t, hid, prod); n != 0 {
		t.Fatalf("want no override row, got %d", n)
	}
}
