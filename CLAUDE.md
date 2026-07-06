# TomoFlix — Instructions for Claude Code

Before making changes in this repo, read these files in order:
1. `context.md` — why decisions were made (stack choices, tradeoffs,
   persistence choice, coding conventions)
2. `ARCHITECTURE.md` — current system shape (what exists right now)
3. `PLAN.md` — what's next, phased and checkbox-tracked

## Working agreement
- This project is also being developed via Codex and via chat sessions
  with Claude (claude.ai). All three should treat `context.md`,
  `ARCHITECTURE.md`, and `PLAN.md` as the shared source of truth, to avoid
  each tool silently diverging on architecture or conventions.
- When you make a significant decision (library choice, schema change,
  new component) or complete a `PLAN.md` phase, update the relevant doc(s)
  in the same session. Don't let the docs drift from the code.
- Code should stay comprehensible to a junior/mid-level developer — favor
  explicit, readable control flow over clever-but-opaque patterns.
  Comment the *why*, especially around concurrency (goroutines, channels).
- Do not add pirate/unlicensed streaming APIs as a video source. See
  `context.md` for the reasoning — this has already been discussed and
  decided against.
