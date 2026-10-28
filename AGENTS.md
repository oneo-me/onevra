# Onevra

Onevra is an independent Go component library built on MyGo. Read [ARCHITECTURE.md](ARCHITECTURE.md) in full before changing code, then the relevant Subjects.

## Principles

- Keep Go code, tests, examples and packaging in `source/`; run Go commands there.
- Use Onevra's components, from `source/components/` and `source/shells/`, instead of MyGo's styled equivalents. Only MyGo component demonstrations are exempt.
- Compose MyGo primitives, unstyled bases and plugins, including Glass, directly. Never build a parallel renderer, input system, theme manager or plugin replacement.
- Keep public components product-neutral. Applications own business state, localized text, persistence and asynchronous work.
- Prefer generic content slots over business-specific fields or variants, and keep existing convenience APIs when adding slots.
- Read MyGo's public signatures before using its APIs. Prefer `docs/` and `ui/` under `GOWORK=off go list -m -f '{{.Dir}}' github.com/egoist/mygo` (from `source/`) over the [latest website](https://mygo.egoist.dev/llms.txt).
- On every MyGo upgrade, follow [Upgrading MyGo](docs/build.md#upgrading-mygo).
- Give every new public component a website page with a live example and `<GoAPI>` entries in both languages; see [Adding a page](website/README.md#adding-a-page).
- Keep Gallery and starter-application trays aligned with the [tray convention](docs/build.md#tray-convention).

## Working with the user

- Never create sub-Agents or Git worktrees; work directly in the current checkout.
- Do not commit, tag, publish or push without explicit authorization.
- An explicit request to publish a named version authorizes [Authorized publication](docs/build.md#authorized-publication) end to end, including the tag-triggered GitHub Release.
- Requests to prepare a release, edit documents or record release rules do not authorize publication. Application archives need separate inclusion.
- Omit verification details from update summaries, changelog entries and release notes unless the user asks. Required checks still run.

## Verification

Scale the [development checks](docs/build.md#development-checks) to the change.

- While iterating, run only affected packages, narrowed with `-run`. Before handing off, run formatting, vet, the build and `go test ./...` once.
- Leave the race check, native tests and runs over a minute to the user; list those commands as not run.
- Website changes run `bun run check`, `bun run typecheck` and `bun run build` in `website/`. Documentation-only changes skip Go checks.
- Test public component contracts and fixed bugs, extending existing tests first; avoid real-time waits and repeated full-page frames.
- Interface changes need target-platform native review; report what remains unreviewed instead of compensating with tests.

## Documentation

- Update a document only when a fact its reader needs changes; link rather than repeat.
- Avoid secondary labels and descriptions that restate a heading or value; let the layout explain.
- Keep each document within its budget and written for its reader. Keep website API doc comments accurate.
- Update `skills/onevra/` when usage, component choices, testing or packaging change.
- Repair documents you touch instead of only appending to them.
- Keep READMEs equivalent. Everything else, including comments, is English except maintained translations and multilingual fixtures.
- Record consumer-visible outcomes under `Unreleased` as one verb-first clause naming the public API, with references such as `(#12 by @alice)` and without a colon, semicolon or "which" clause.
- Omit refactors, layout, tests, tooling, styling, copy and gallery-only changes.
- Prefix breaking changes with `Breaking:`, naming at most the replacement. Record a MyGo upgrade as the required version alone, first.
- Order each version's list as `Breaking:`, Add, other, Fix; merge related entries.
- Keep documentation on the code branch and reconcile it during merges.
- Check budgets, Subjects, links, anchors and changelog structure; run `git diff --check`.
