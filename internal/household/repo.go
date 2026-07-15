package household

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func (r *Repo) HouseholdMembers(ctx context.Context, hid int) ([]Member, error) {
	sql := `
	SELECT hm.user_id, u.username, hm.role 
	FROM household_members hm
	JOIN users u ON u.id=hm.user_id
	WHERE hm.household_id=$1;
	`
	rows, err := r.db.Query(ctx, sql, hid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		var role string
		if err := rows.Scan(&m.ID, &m.Name, &role); err != nil {
			return nil, err
		}
		m.IsOwner = role == "owner"
		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *Repo) Household(ctx context.Context, hid int) (Household, error) {
	sql := `
	SELECT id, name, code 
	FROM households
	WHERE id=$1;
	`
	var household Household
	err := r.db.QueryRow(ctx, sql, hid).Scan(&household.ID, &household.Name, &household.Code)
	if err != nil {
		return Household{}, err
	}

	return household, nil
}

func NewRepo(db *pgxpool.Pool) *Repo {
	if db == nil {
		panic("nil DB connection")
	}
	return &Repo{db: db}
}

func (r *Repo) RegenerateHouseholdCode(ctx context.Context, uid int, code string, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var callerRole string
	err = tx.QueryRow(ctx, `SELECT role FROM household_members WHERE user_id=$1 AND household_id=$2;`, uid, hid).Scan(&callerRole)
	if err != nil {
		return fmt.Errorf("caller not found in household: %w", err)
	}
	if callerRole != "owner" {
		return fmt.Errorf("caller is not an owner")
	}

	_, err = tx.Exec(ctx, `UPDATE households SET code=$1 WHERE id=$2;`, code, hid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) RemoveMember(ctx context.Context, userID, targetID int, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
	SELECT role
	FROM household_members
	WHERE user_id=$1 AND household_id=$2;
	`
	var callerRole string
	err = tx.QueryRow(ctx, sql, userID, hid).Scan(&callerRole)
	if err != nil {
		return fmt.Errorf("caller not found in household: %w", err)
	}
	if callerRole != "owner" {
		return fmt.Errorf("caller is not an owner")
	}

	sql = `
	DELETE
	FROM household_members
	WHERE user_id=$1 and household_id=$2;
	`
	_, err = tx.Exec(ctx, sql, targetID, hid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) PromoteMember(ctx context.Context, uid int, targetUID int, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
	SELECT role
	FROM household_members
	WHERE user_id=$1 AND household_id=$2;
	`
	var callerRole string
	err = tx.QueryRow(ctx, sql, uid, hid).Scan(&callerRole)
	if err != nil {
		return fmt.Errorf("caller not found in household: %w", err)
	}
	if callerRole != "owner" {
		return fmt.Errorf("caller is not an owner")
	}

	sql = `
	SELECT EXISTS(
		SELECT 1
		FROM household_members
		WHERE user_id=$1 AND household_id=$2
	);
	`
	var targetExists bool
	err = tx.QueryRow(ctx, sql, targetUID, hid).Scan(&targetExists)
	if err != nil {
		return err
	}
	if !targetExists {
		return fmt.Errorf("target user not found in household")
	}

	sql = `
	UPDATE household_members
	SET role='owner'
	WHERE user_id=$1 AND household_id=$2;
	`
	_, err = tx.Exec(ctx, sql, targetUID, hid)
	if err != nil {
		return err
	}

	sql = `
	UPDATE household_members
	SET role='member'
	WHERE user_id=$1 AND household_id=$2;
	`
	_, err = tx.Exec(ctx, sql, uid, hid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) HouseholdVersion(ctx context.Context, householdID int) (int, error) {
	sql := `
	SELECT version
	FROM households
	WHERE id = $1;
	`

	var version int
	err := r.db.QueryRow(ctx, sql, householdID).Scan(&version)

	return version, err
}

func (r *Repo) CreateInvite(ctx context.Context, householdID, createdBy int, expiresAt time.Time) (string, error) {
	sql := `
	INSERT INTO invites (household_id, created_by, expires_at)
	VALUES ($1, $2, $3)
	RETURNING token;
	`

	var token string
	err := r.db.QueryRow(ctx, sql, householdID, createdBy, expiresAt).Scan(&token)

	return token, err
}

func (r *Repo) CurrentInvites(ctx context.Context, householdID int) ([]Invite, error) {
	sql := `
	SELECT token, created_at, expires_at
	FROM invites
	WHERE household_id = $1 AND expires_at > now()
	ORDER BY created_at DESC;
	`

	rows, err := r.db.Query(ctx, sql, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []Invite
	for rows.Next() {
		var inv Invite
		if err := rows.Scan(&inv.Token, &inv.CreatedAt, &inv.ExpiresAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}

	return invites, rows.Err()
}

func (r *Repo) RemoveExpiredInvites(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM invites WHERE expires_at <= now();`)
	return err
}

func (r *Repo) RevokeInvite(ctx context.Context, token string, householdID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM invites WHERE token=$1 AND household_id=$2;`, token, householdID)
	return err
}

// deleteHouseholdData removes a household and everything scoped to it. Tables
// that reference households(id) without ON DELETE CASCADE must be cleared first.
func deleteHouseholdData(ctx context.Context, tx pgx.Tx, hid int) error {
	stmts := []string{
		`DELETE FROM groceries_history WHERE household_id=$1;`,
		`DELETE FROM grocery_lists WHERE household_id=$1;`,
		`DELETE FROM recipes WHERE household_id=$1;`,
		`DELETE FROM restaurants WHERE household_id=$1;`,
		`DELETE FROM invites WHERE household_id=$1;`,
		`DELETE FROM household_members WHERE household_id=$1;`,
		`DELETE FROM households WHERE id=$1;`,
	}
	for _, sql := range stmts {
		if _, err := tx.Exec(ctx, sql, hid); err != nil {
			return err
		}
	}
	return nil
}

// LeaveHousehold removes a non-owning member from their household. Owners must
// transfer ownership (or delete the household) instead, so the household is
// never left without an owner.
func (r *Repo) LeaveHousehold(ctx context.Context, uid, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var role string
	err = tx.QueryRow(ctx, `SELECT role FROM household_members WHERE user_id=$1 AND household_id=$2;`, uid, hid).Scan(&role)
	if err != nil {
		return fmt.Errorf("caller not found in household: %w", err)
	}
	if role == "owner" {
		return ErrOwnerMustTransfer
	}

	_, err = tx.Exec(ctx, `DELETE FROM household_members WHERE user_id=$1 AND household_id=$2;`, uid, hid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteHousehold deletes the household and all its data. Only the owner may do
// this, and only when no other members remain.
func (r *Repo) DeleteHousehold(ctx context.Context, uid, hid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var role string
	err = tx.QueryRow(ctx, `SELECT role FROM household_members WHERE user_id=$1 AND household_id=$2;`, uid, hid).Scan(&role)
	if err != nil {
		return fmt.Errorf("caller not found in household: %w", err)
	}
	if role != "owner" {
		return ErrOwnerMustTransfer
	}

	var others int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM household_members WHERE household_id=$1 AND user_id != $2;`, hid, uid).Scan(&others)
	if err != nil {
		return err
	}
	if others > 0 {
		return ErrHouseholdNotEmpty
	}

	if err := deleteHouseholdData(ctx, tx, hid); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteAccount permanently removes a user. If the user solely owns a household
// it is deleted with them; if they own one with other members they must transfer
// ownership first. Households they created but no longer own are repointed to the
// current owner so the created_by reference stays valid.
func (r *Repo) DeleteAccount(ctx context.Context, uid int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT household_id, role FROM household_members WHERE user_id=$1;`, uid)
	if err != nil {
		return err
	}
	type membership struct {
		hid   int
		owner bool
	}
	var memberships []membership
	for rows.Next() {
		var m membership
		var role string
		if err := rows.Scan(&m.hid, &role); err != nil {
			rows.Close()
			return err
		}
		m.owner = role == "owner"
		memberships = append(memberships, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// Owners with company can't vanish without orphaning the household.
	for _, m := range memberships {
		if !m.owner {
			continue
		}
		var others int
		err = tx.QueryRow(ctx, `SELECT count(*) FROM household_members WHERE household_id=$1 AND user_id != $2;`, m.hid, uid).Scan(&others)
		if err != nil {
			return err
		}
		if others > 0 {
			return ErrOwnerMustTransfer
		}
	}

	for _, m := range memberships {
		if m.owner {
			if err := deleteHouseholdData(ctx, tx, m.hid); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.Exec(ctx, `DELETE FROM household_members WHERE user_id=$1 AND household_id=$2;`, uid, m.hid); err != nil {
			return err
		}
	}

	// Hand any households this user created (but no longer owns) to their owner.
	_, err = tx.Exec(ctx, `
	UPDATE households h
	SET created_by = (
		SELECT user_id FROM household_members
		WHERE household_id = h.id AND role = 'owner'
		LIMIT 1
	)
	WHERE created_by = $1;`, uid)
	if err != nil {
		return err
	}

	for _, sql := range []string{
		`DELETE FROM invites WHERE created_by=$1;`,
		`DELETE FROM sessions WHERE user_id=$1;`,
		`DELETE FROM users WHERE id=$1;`,
	} {
		if _, err := tx.Exec(ctx, sql, uid); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
