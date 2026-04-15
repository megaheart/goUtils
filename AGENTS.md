# AGENTS.md

## Repository snapshot
- Go module: `github.com/megaheart/goUtils` (`go.mod` declares Go `1.23.3`).
- This is a utility library repo, not a single runnable app.
- Directory is code-heavy (not empty / not config-only).

## Essential commands
Use these from repo root:

```bash
go test ./...
```
- Primary validation command currently in use.
- At time of writing, only `./test` has test files; other packages report `[no test files]`.

No Makefile, no CI workflow files, and no dedicated lint script were found.

## Existing instruction files
Checked and not found:
- `.cursor/rules/*.md`
- `.cursorrules`
- `.github/copilot-instructions.md`
- `claude.md`
- `agents.md` (lowercase)

This file (`AGENTS.md`) is the first repo-level agent instruction file.

## Code organization
### Root package (`package goUtils`)
Core utility types and algorithms live at root:
- `primitiveSet.go`: sorted generic set for primitive comparable types (`PrimitiveSet[T]`) using binary search and `slices.Insert/Delete`.
- `set.go`: sorted generic set with caller-supplied comparator (`Set[T]`).
- `primitiveReadOnlyMap.go`: read-only key/value map backed by sorted keys + aligned values arrays.
- `heapSort.go`: `HeapArray` with in-place heap sort.
- `collectionHelper.go`: random-pick helper (`PickRandom`).
- `time.go`: custom JSON-unmarshalable date/time wrapper types (`DateTime`, `Date`, `TimeInDay`, `Weekday`).
- `tuple.go`: generic `Tuple2`.

### Subpackages
- `fs/`: filesystem abstraction layer and implementations.
  - `interfaces.go`: `IFile`, `IFileWatcher`, `IFileSystem`, plus file-op constants.
  - `osFileSystem.go`: OS-backed implementation (`OsFileSystem`, `OsFileWatcher`).
  - `aferoFileSystem.go`: in-memory implementation based on `afero` (`AferoFileSystem`, `AferoFileWatcher`).
  - `utils/watchFile.go` (`package fsUtils`): higher-level debounced file watching + JSON file helpers.
- `log/`: logging abstraction (`ILogger`, `LogField`) + zap/lumberjack implementation (`ZapLogger`).
- `event/`: simple event manager (`EventManager`, generic `DataEventManager[T]`).
- `appHelper/`: reflection-driven Cobra parameter parsing from struct tags.
- `test/`: table-driven tests for `HeapArray` and `PrimitiveSet`.
- `test_exec/`: standalone executable used as manual/randomness experiment harness.

## Architectural patterns and data flow
### 1) Sorted-array collections are the default pattern
- Both `PrimitiveSet` and `Set` maintain sorted slices and rely on binary-search-style lookups.
- Mutation operations preserve order by insertion at computed index.
- `PrimitiveReadOnlyMap` depends on sorted keys + index alignment into value slice.

Implication: if extending these types, preserve sorted invariants first; many methods assume sorted data.

### 2) Interface-first filesystem abstraction
- `fs.IFileSystem` is broad and acts as a common contract across OS and in-memory backends.
- Higher-level watcher code (`fs/utils/watchFile.go`) consumes the interface, not concrete types.

Implication: new behavior should usually be added to both `OsFileSystem` and `AferoFileSystem` to keep parity.

### 3) Logging decouples call sites from zap
- App code should use `log.ILogger` and `log.LogField` helpers.
- Zap-specific behavior is isolated in `log/zapLogger.go`.

## Non-obvious gotchas (important)
1. `AferoFileSystem.Symlink` is intentionally unsupported for default mem-fs and returns `*os.LinkError`.
   - If tests require symlinks, use `OsFileSystem` or switch underlying fs strategy.

2. `AferoFileSystem.IsDirError` and `AferoFileSystem.IsNotDirError` currently `panic("...not implemented")`.
   - Avoid calling these in generic flows unless you implement them first.

3. `fs/utils/FileWatcher` is explicitly documented as **NOT THREAD-SAFE**.
   - Do not treat registration/state mutation APIs as concurrent-safe.

4. `WatchFile` immediately executes `onChange()` once before watcher registration.
   - Side effects happen at registration time, not only on file changes.

5. In `appHelper.ParseParamsAndSubCommands`, env vars override default field values before flags are bound; comment text in file is inconsistent about priority.
   - Validate precedence behavior with tests before relying on docs/comments.

6. Tests use `package test` under `/test` and call into module packages via imports (`github.com/megaheart/goUtils`).
   - Add new tests in `/test` unless you intentionally want package-internal tests.

## Testing patterns
- Table-driven tests with subtests (`t.Run`) and `t.Parallel()` are used.
- `test/primitiveSet_test.go` includes a local helper `LimitTestTime(...)` that wraps test logic with a timeout channel.
- Assertions use `stretchr/testify/assert`.

When adding tests, match this style unless you have a strong reason not to.

## Naming and style conventions observed
- Exported types/functions use Go naming conventions.
- Generic constructors follow `NewX`, `NewXWith...` patterns.
- Many files include verbose doc comments, sometimes with examples and complexity notes.
- Internal field naming often uses simple names (`Array`, `key`, `value`, `compare`).
- `goUtils` (camel-case package name) is used as the root package identifier.

## Practical workflow for future agents
1. Make targeted changes in the relevant package(s).
2. If you touch `fs` contracts, check both OS and Afero implementations for parity.
3. Run:
   ```bash
   go test ./...
   ```
4. For collection changes, include edge cases for empty/singleton/even/odd-length inputs (existing tests already follow this heavily).

## Scope boundaries in this repo
- No deployment pipeline/config was found.
- No CLI entrypoint for production app behavior was found (only `test_exec/` demo executable).
- This repository is primarily reusable library code.
