---
format: https://specscore.md/feature-specification
status: Draft
---

# Feature: Telegram Reversi bot (vs AI, state in callback data)

> [SpecScore.**Studio**](https://specscore.studio): | [Explore](https://specscore.studio/app/github.com/prizarena/reversi/spec/features/telegram-reversi-bot?op=explore) | [Edit](https://specscore.studio/app/github.com/prizarena/reversi/spec/features/telegram-reversi-bot?op=edit) | [Ask question](https://specscore.studio/app/github.com/prizarena/reversi/spec/features/telegram-reversi-bot?op=ask) | [Request change](https://specscore.studio/app/github.com/prizarena/reversi/spec/features/telegram-reversi-bot?op=request-change) |
**Status:** Draft
**Source Ideas:** reversi-as-a-sneat-bot-game

## Summary

Play **Reversi** (Othello) against a built-in robotic opponent inside a single
Telegram chat, in one of two modes — **AI** (a heuristic player) or **Random**.
The board is an 8×8 inline keyboard; the **complete game state is carried in each
button's callback data**, so there is no server-side game storage. The game rules
and AI come from this repo's existing `revgame` engine — the bot layer holds no
game logic. v1 is single-player vs the chosen opponent; the play surface is
designed to be reused by any Sneat bot (first consumer: SneatBot's `/games`).

This document follows the [SpecScore feature specification](https://specscore.md/feature-specification).

## Problem

Reversi already exists here as a standalone Google App Engine app built on
`strongo/bots-framework` — the **predecessor** of today's
`github.com/bots-go-framework/bots-fw`, the same framework the current Sneat bots
run on. Because it's the same lineage, migrating the old wiring forward is mostly
mechanical (import/type renames like `bots.MessageFromBot` → `botmsg.MessageFromBot`,
`bots.Command` → `botsfw.Command`). Its **engine** (`server-go/revgame`: `Board`,
`Disks` bitboards, `SimpleAI`, base64 int/transcript encoding) is solid,
test-covered, and reused directly; its **bot wiring** (`server-go/revbot`) is
migrated, dropping only the genuinely dead/copy-pasted bits (e.g. `render_board.go`,
which is mostly commented-out code referencing another game's models).

The one thing that genuinely changes is the **state model**: the old app stored
boards in a `strongo/db` datastore, whereas v1 carries the whole board state in
callback data — removing persistence, "resume", and concurrency-guard surface
entirely. The game *is* the message, so any number of games can run at once and an
old game message stays playable when scrolled back to. The play layer is built so
any Sneat bot can mount it.

## Behavior

### Engine reuse

#### REQ: engine-is-single-source-of-rules

All Reversi rules — legal-move generation, disk flipping, scoring, game-over, and
the AI opponent — MUST come from the `revgame` engine (`Board.MakeMove`,
`SimpleAI.GetMove`, etc.). The bot/rendering layer MUST NOT re-implement or
duplicate any game rule.

#### REQ: current-bots-framework

The bot integration MUST target the current **`bots-go-framework`** modules used
by the Sneat bots — `github.com/bots-go-framework/bots-fw` (`botsfw`, `botmsg`,
`botinput`), `github.com/bots-go-framework/bots-go-core/botkb` (keyboards), and
`github.com/bots-go-framework/bots-fw-telegram` — at the versions SneatBot
currently pins (see `sneat-co/sneat-go` `go.mod`; at time of writing `bots-fw`
v0.75.18, `bots-go-core` v0.2.5, `bots-fw-telegram` v0.27.6). This is the current
name/version of the framework the old app already used (`strongo/bots-framework`),
so it is a migration, not a rewrite; the shipped code MUST NOT still import the old
`github.com/strongo/bots-framework` / `github.com/strongo/db` paths. Rendered
keyboards MUST be built with `botkb`.

### Opponent modes

#### REQ: opponent-modes

The human MUST be able to play against one of two robotic opponents, chosen when
starting a new game:

- **AI** — the existing `revgame.SimpleAI`, used **as-is** (greedy score-max with a
  corner preference). It breaks ties with randomness, so it is intentionally *not*
  reproducible in v1 — that is accepted.
- **Random** — picks a uniformly random legal move.

Both modes MUST always return a legal move when one exists (and are only consulted
when the opponent has a legal move — see REQ:auto-pass). The chosen mode MUST
persist for the whole game so every reply move uses the same opponent.

> Note (not in v1): a stronger, *deterministic* opponent is a documented future
> option — a pure-Go negamax + alpha-beta with a weighted-square evaluation table,
> made reproducible by a fixed-order tie-break (MIT reference:
> `github.com/ReconGit/go-othello-ai`). Strong C/C++ engines (Edax GPL-2.0,
> Egaroucid GPL-3.0) were reviewed and rejected as overkill and licensing/toolchain
> baggage for a casual in-chat game.

### Board rendering

#### REQ: board-as-inline-keyboard

The board MUST render as an 8×8 grid of inline-keyboard buttons — one button per
cell — showing black disks, white disks, and empty cells, so a move is made by
tapping a cell rather than typing coordinates.

#### REQ: start-position

Starting a new game MUST let the human pick the opponent mode (AI or Random), then
render the Othello starting position: black disks at d5 and e4, white disks at d4
and e5, with the human playing Black and moving first.

### State in callback data

#### REQ: state-in-callback-data

Every board button's `callback_data` MUST encode the complete board snapshot —
the black bitboard, the white bitboard, the last move, and the opponent mode
(AI / Random) — plus the cell that button targets, using the engine's base64
encoder (`EncodeIntToBase64`). The total `callback_data` MUST stay within
Telegram's 64-byte limit for every reachable board. The snapshot MUST round-trip:
decoding a button's data MUST reproduce the exact `Board` and opponent mode it was
rendered from.

#### REQ: no-server-persistence

The bot MUST NOT persist game state server-side (no datastore/Firestore records,
no per-chat "current game" blob). Board state lives only in the rendered message's
callback data. Accepted v1 consequences: no move history survives beyond the
current snapshot, and any user who can tap the message advances the move (no
per-player lock on a shared or forwarded message).

### Playing a move

#### REQ: human-move-then-opponent

Tapping a legal empty cell MUST apply the human's Black move via the engine
(placing the disk and flipping all flanked white disks), then let the selected
opponent (AI or Random, per the snapshot's mode) make exactly one White reply
move, then re-render the updated board in place (edited message) with fresh
callback data on every button.

#### REQ: illegal-move-noop

Tapping a cell that is occupied or is not a legal move MUST leave the board state
unchanged and re-render the same board (optionally with a brief "not a legal move"
note). No opponent move is triggered.

#### REQ: auto-pass

When the side to move has no legal move but the game is not over, that side's
turn MUST be skipped automatically (a pass) rather than stalling the game.

#### REQ: game-over

When neither side has a legal move (or the board is full), the board MUST render
as finished — showing each side's final disk count and the winner (or a draw) —
with a button to start a new game.

## Architecture & Components

- **`revgame` engine** (existing, in `server-go/revgame`) — kept as the sole
  source of rules + AI. Its public API used by the play layer: `Board`,
  `Board.MakeMove`, valid-move / score / game-over queries, `SimpleAI.GetMove`,
  and `EncodeIntToBase64` / `DecodeIntFromBase64`.

- **Reversi play layer** (this repo) — a package that (a) renders a `Board` to a
  `bots-go-core/botkb` inline keyboard, and (b) encodes/decodes the callback-data
  snapshot. It depends on `revgame` plus the current `bots-go-framework` keyboard
  module (`botkb`) — the same keyboard abstraction every current Sneat bot renders
  through — and **not** on any messenger API directly, so it stays portable across
  the framework's platform adapters. It MUST NOT pull in the legacy
  `strongo/bots-framework`.

- **Host-bot wiring** (out of this repo — e.g. SneatBot in `sneat-co/sneat-go`)
  registers the `botsfw.Command`, calls the play layer, and returns its
  `botmsg.MessageFromBot`. Reusable parts of `server-go/revbot` (board rendering,
  move handling) are **migrated** from `strongo/bots-framework` to `bots-fw` — a
  mostly-mechanical port — rather than rewritten from scratch; only the genuinely
  dead bits are dropped, and the datastore-backed state model is replaced by
  callback data. (Note: the repo's own `go.mod` still targets the predecessor
  modules today — bumping to the current `bots-go-framework` modules is part of
  building this feature.)

**Callback-data layout** (illustrative): `<b64(blacks)>.<b64(whites)>.<b64(last)>`
is the snapshot; a button appends its target cell index (0–63), e.g.
`…&c=<idx>`. Worst case ≈ two 11-char bitboards + separators + a short index,
comfortably under 64 bytes.

**Data flow:** tap → decode snapshot from `callback_data` → `Board.MakeMove`
(human) → `SimpleAI.GetMove` → `Board.MakeMove` (AI) → auto-pass check → encode
the new snapshot into every button → edit the message.

## Testing strategy

- Keep `revgame`'s existing engine tests green (rules + AI already covered).
- Add a callback-data round-trip test: encode → decode reproduces the `Board`
  and the encoded string is ≤64 bytes for boundary boards (empty, full, and a
  board with a set last move).
- Add play-layer tests: a new game renders the start position; a legal tap flips
  the flanked disks and produces exactly one opponent reply; an illegal tap is a
  no-op; a no-legal-move side auto-passes; a finished board renders the result.
- Opponent-mode tests: Random returns only legal moves and the mode survives the
  callback-data round-trip; AI mode routes to `SimpleAI`.

## Not Doing / Out of Scope (v1)

- Player-vs-player (two Telegram users), invites, matchmaking, cross-chat turn
  notifications.
- Move history / transcript beyond the current snapshot; "resume game" / "my
  games" lists.
- Per-player lock on a shared or forwarded game message.
- Other messengers (Viber, etc.) — the play layer is framework-light to allow it
  later, but only Telegram is in scope now.
- Additional difficulty tiers beyond the two v1 opponent modes (AI + Random). A
  stronger deterministic engine may land later but is not required for v1.

## Acceptance Criteria

### AC: rules-come-from-engine (verifies REQ:engine-is-single-source-of-rules)

**Given** the Reversi bot play layer
**When** a move is validated, applied, scored, or an AI reply is computed
**Then** it is done by calling the `revgame` engine, and the play layer contains
no independent move-generation, flipping, or scoring logic.

### AC: uses-current-framework-not-legacy (verifies REQ:current-bots-framework)

**Given** the built Reversi bot integration and its module graph
**When** its dependencies are inspected
**Then** it imports `bots-go-framework` modules (`bots-fw`, `bots-go-core/botkb`,
`bots-fw-telegram`) and does not import `github.com/strongo/bots-framework` or
`github.com/strongo/db`.

### AC: board-renders-8x8-buttons (verifies REQ:board-as-inline-keyboard)

**Given** any `Board`
**When** it is rendered for Telegram
**Then** the inline keyboard has 8 rows of 8 cell buttons, each cell showing a
black disk, a white disk, or an empty marker.

### AC: new-game-start-position (verifies REQ:start-position)

**Given** a request to start a new Reversi game and a chosen opponent mode
**When** the new-game action runs
**Then** the board shows black at d5/e4 and white at d4/e5, it is the human's
(Black) turn, and the chosen opponent mode is recorded in the board's callback data.

### AC: opponent-modes-selectable (verifies REQ:opponent-modes)

**Given** the human chooses the Random opponent (and separately, the AI opponent)
**When** it is the opponent's turn and at least one legal move exists
**Then** Random returns one uniformly-random legal move and AI returns
`SimpleAI`'s move, and in both cases the returned move is legal.

### AC: callback-data-round-trips-within-limit (verifies REQ:state-in-callback-data)

**Given** any reachable board snapshot (both bitboards, last move, and opponent mode)
**When** it is encoded into a button's `callback_data` and then decoded
**Then** the decoded `Board` and opponent mode equal the originals and the encoded
`callback_data` is at most 64 bytes.

### AC: no-records-written (verifies REQ:no-server-persistence)

**Given** a full Reversi game played from start to finish in a chat
**When** the game runs
**Then** no game-state records are written to any datastore and no per-chat game
blob is stored — the state exists only in the message callback data.

### AC: legal-tap-flips-and-triggers-opponent (verifies REQ:human-move-then-opponent)

**Given** a board where it is the human's turn and cell X is a legal Black move
that flanks at least one white disk
**When** the human taps cell X
**Then** cell X becomes Black, every flanked white disk on that line flips to
Black, and exactly one opponent (White) move is applied before the board is
re-rendered.

### AC: illegal-tap-is-noop (verifies REQ:illegal-move-noop)

**Given** a rendered board and a cell Y that is occupied or not a legal move
**When** the human taps cell Y
**Then** the board state is unchanged and no opponent move is made.

### AC: no-legal-move-auto-passes (verifies REQ:auto-pass)

**Given** a board where the side to move has no legal move but the opponent does
**When** the game advances
**Then** that side's turn is skipped and play continues with the opponent.

### AC: finished-board-shows-result (verifies REQ:game-over)

**Given** a board where neither side has a legal move
**When** the board is rendered
**Then** it shows each side's final disk count and the winner (or draw), plus a
"new game" button.

## Open Questions

None at this time.

---
*This document follows the https://specscore.md/feature-specification*
