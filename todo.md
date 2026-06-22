# TODO

## Invite-only account creation + household management

Goal: lock down account creation so it's **invite-only**, and add a profile/settings
page for managing a household. (Builds on the existing `GET /register` page, which is
currently just the view — `POST /register` is not implemented yet.)

### Invite-only registration
- [ ] **Make account creation invite-only** — no open self-signup. A new account can
  only be created via a valid invite.
- [ ] **Any logged-in member can generate an invite link** — produces a link (with a
  token) that lets exactly one new person create an account.
- [ ] **Invite link → account** — opening the link leads to the register page; on submit
  it creates the user and adds them to the inviter's household.
- [ ] Decide invite semantics: single-use vs. reusable, expiry, and whether to also
  support a short **household code** (typed in) in addition to the full link.

### Profile / settings page
- [ ] **New profile/settings page** (behind auth) where a member can:
  - [ ] **Get the household code** for their household.
  - [ ] **Retrieve / generate the invite link** to create an account.
  - [ ] **Remove users from the household** — owner only.
  - [ ] **Transfer ownership** to another member.

### Open questions / decisions
- [ ] Introduce a household **owner** concept (no owner/role exists today — members just
  share a household). Needed for "remove user" and "transfer ownership".
- [ ] Where invites/codes live (new table? on the household?) and how tokens are
  generated/validated/expired.
- [ ] What "remove a user" does to their data and whether a removed user can be deleted
  entirely or just detached from the household.
