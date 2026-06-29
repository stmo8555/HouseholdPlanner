# Ticket: Admin Dashboard

## Goal

Add a platform-level admin who logs in and lands directly on an admin dashboard
(no household required) to oversee the whole system.

## Background

Today roles only exist inside a household (`owner`/`member`), and every
authenticated route except onboarding is gated behind `RequireHousehold`. There
is no concept of a system administrator.

## Scope

- **Admin identity**: a `users.is_admin` boolean column. The admin account is
  inserted by a DB startup script with `is_admin = TRUE`. On login, admins are
  redirected to `/admin` instead of `/groceries`.
- **View** all households with their members and current invite links.
- **Member roles**: change a member's role (`owner`/`member`) within a household,
  bypassing the normal owner-only rules.
- **Users**: create users (credentials only — username + password), clear a
  user's sessions (force logout), change a user's password, and delete a user.
- **Households**: delete a household and all of its data (cascade).
- **Invitations**: send out admin invitation links via a new `admin_invites`
  table (not tied to any household). Opening the link lets a new person register
  and then create/join their own household via the existing onboarding flow.
- **Dashboard stats**: counts of users, households, active sessions, and
  outstanding invites.
- **Revoke invite links**: kill a household or admin invite link on demand.

## Constraints / decisions

- Access is guarded by a new `RequireAdmin` middleware; non-admins hitting
  `/admin` are rejected, and admins hitting household routes are bounced to
  `/admin`.
- No email sending — invite links are copyable.
- No migration framework — `init.sql` is edited directly and the DB regenerated
  (no production DB exists yet).
- Follows the existing Handler → Service → Repo pattern in a new `internal/admin`
  package; reuses existing cascade-delete logic in `internal/household`.

## Out of scope

Email delivery of invites, an audit log, and impersonation / "log in as".
