---
format: https://specscore.md/idea-specification
status: Specifying
---

# Idea: Reversi as a Sneat bot game

**Status:** Specifying
**Date:** 2026-07-23
**Owner:** alex
**Promotes To:** telegram-reversi-bot
**Supersedes:** —
**Related Ideas:** —

## Problem Statement

How might we bring Reversi back to life as a zero-maintenance, no-persistence game that lives inside Sneat's messengers?

## Context

Reversi (Othello) shipped years ago as a standalone Google App Engine app on
`strongo/bots-framework` (`@ReversiGameBot`) — the **predecessor** of today's
`bots-go-framework/bots-fw`, so the old wiring migrates forward with mostly
mechanical import/type renames. The **engine** (`server-go/revgame`: bitboard
`Board`, `SimpleAI`, base64 encoders) is small, dependency-light, and test-covered,
and reused directly; the **bot wiring** (`server-go/revbot`) is migrated, dropping
only some genuinely dead / copy-pasted bits. Meanwhile Sneat's current bots have no
play surface at all, and the ecosystem is starting a `/games` menu in SneatBot.
Reversi is the natural first game because the hard part (a correct engine + AI)
already exists.

## Recommended Direction

Revive Reversi **engine-first**: keep `revgame` as the single source of rules and
AI, and build a fresh, framework-light play layer on top of it (render a board to
an abstract keyboard + encode/decode game state). Mount that play layer in the
Sneat bots — SneatBot's `/games` being the first consumer — instead of resurrecting
the old GAE app or its `revbot` wiring.

The defining choice is **state in callback data, not in a database**: every board
button carries the full board snapshot (both bitboards + last move) in its
`callback_data`, base64-encoded, under Telegram's 64-byte limit. This deletes the
entire persistence / "resume" / concurrency-guard surface — the game *is* the
message. Any number of games can run at once, and a scrolled-back game message stays
playable. v1 is single-player, with two selectable robotic opponents: **AI** (the
existing `SimpleAI`, kept as-is) and **Random** (a uniformly random legal move).

## Alternatives Considered

- **Resurrect the GAE app + `revbot` as-is** — lost: keeps a standalone app and its
  `strongo/db` datastore-backed state model, when we want callback-data state
  mounted inside the Sneat bots. (The framework itself is fine — it's the ancestor
  of `bots-fw` — so we migrate its reusable wiring rather than resurrect the app.)
- **Store games in Firestore/dal-go** — lost for v1: adds persistence, a "resume"
  surface, and per-player locking for a casual game, when callback data removes all
  of it. Can be added later if PvP/history is ever wanted.
- **Player-vs-player first** — lost: needs invites, cross-chat turn notifications,
  and turn-ownership tracking — most of where the legacy bot's complexity came from.
  Vs-AI reuses the existing engine with none of that.

## MVP Scope

A player can open Reversi in a Telegram chat, pick an opponent (AI or Random), see
the Othello start position as a tappable 8×8 board, play a full game to its end,
and start a new one — with **no game state stored anywhere but the message itself**.

## Not Doing (and Why)

- Player-vs-player, invites, matchmaking, cross-chat turn notifications — deferred; where the legacy bot's complexity lived, and vs-AI needs none of it.
- Move history / transcript, "resume game" / "my games" lists — a growing transcript will not fit in callback data; out of scope for a stateless v1.
- Per-player lock on a shared/forwarded message — accepted v1 trade-off of state-in-callback-data.
- Other messengers (Viber, etc.) — the play layer stays framework-light to allow it later, but only Telegram is in scope now.
- A stronger / deterministic AI engine — deferred. A small pure-Go negamax +
  alpha-beta with a weighted-square table (deterministic via fixed-order tie-break;
  MIT reference `github.com/ReconGit/go-othello-ai`) is the documented future path;
  strong C/C++ engines (Edax, Egaroucid) were reviewed and rejected as overkill and
  GPL/toolchain baggage. v1 keeps `SimpleAI` as-is (accepting its random tie-break).

## Key Assumptions to Validate

| Tier | Assumption | How to validate |
|------|------------|-----------------|
| Must-be-true | The full board state (2×int64 + last move) base64-encodes into `callback_data` within Telegram's 64-byte limit for every reachable board. | Round-trip test on boundary boards (empty, full, set last move); measure encoded length. |
| Must-be-true | `revgame` ports/builds cleanly under the current Go toolchain and is the only rules source the bot needs. | Compile + run the existing `revgame` tests in the target module. |
| Should-be-true | The existing `SimpleAI` is a good-enough opponent for a casual in-chat game. | Play-test a few full games; confirm it plays sensibly (takes corners, flips well). |
| Might-be-true | Players accept the v1 trade-offs (no history, no per-player lock on a shared message). | Ship to `/games`, watch for confusion/complaints. |

## SpecScore Integration

- **New Features this would create:** [`telegram-reversi-bot`](../features/telegram-reversi-bot/README.md)
- **Existing Features affected:** none
- **Dependencies:** the `revgame` engine (this repo); host-bot wiring in `sneat-co/sneat-go` (SneatBot `/games`) and the ecosystem `games` feature in `sneat-co/backstage`.

## Open Questions

None at this time.
