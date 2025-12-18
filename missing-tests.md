### Issue: Add missing unit tests for `internal/models.ValidateCoordinates`
**Why**
`ValidateCoordinates` is used by both the CLI (`cmd/position/add.go`) and MCP tool handler (`internal/mcp/tools.go`). It currently has no direct tests, and it’s a shared validation boundary.

**Add tests**
- `TestValidateCoordinates_ValidEdgeValues`
  - lat = -90, 90; lng = -180, 180 should succeed
- `TestValidateCoordinates_InvalidLatitudeTooLow/TooHigh`
  - lat = -90.0001, 90.0001 should error
- `TestValidateCoordinates_InvalidLongitudeTooLow/TooHigh`
  - lng = -180.0001, 180.0001 should error
- `TestValidateCoordinates_ErrorMessages`
  - assert error strings match current implementation (“latitude must be…”, “longitude must be…”), since CLI/MCP return these directly

---

### Issue: Add missing DB tests for uniqueness + error behavior in `internal/db/items.go`
**Why**
`items.name` is `UNIQUE NOT NULL`. Current tests don’t cover duplicates or verify error types/messages. Also `GetItemByID` exists but isn’t tested.

**Add tests**
- `TestCreateItem_DuplicateNameFails`
  - create two items with same name (different UUIDs); second `CreateItem` should error
- `TestGetItemByID_Success`
  - create item, fetch by ID, verify fields
- `TestGetItemByID_NotFoundReturnsError`
  - random UUID should return error
- `TestDeleteItem_NotFoundReturnsError`
  - delete random UUID should return “item not found” error (current behavior)

---

### Issue: Add missing DB tests for foreign key enforcement and invalid references in `internal/db/positions.go`
**Why**
Schema includes FK with `ON DELETE CASCADE`, and `InitDB` enables `PRAGMA foreign_keys=ON`. Tests currently only validate cascade via deleting an item, but don’t prove FK enforcement is actually active.

**Add tests**
- `TestCreatePosition_FailsForUnknownItemID`
  - call `CreatePosition` with `ItemID` not present in `items`; should error (FK constraint)
- `TestInitDB_EnablesForeignKeys`
  - after `InitDB`, run `PRAGMA foreign_keys;` and assert it returns `1`

---

### Issue: Add missing DB tests for ordering and tie-breaking in `GetCurrentPosition` / `GetTimeline`
**Why**
Both queries order only by `recorded_at DESC`. If two positions have identical `recorded_at`, ordering is undefined, and “current” may be non-deterministic.

**Add tests**
- `TestGetTimeline_OrderIsNewestFirstByRecordedAt`
  - already partially tested, but should assert full ordering across multiple rows deterministically
- `TestGetCurrentPosition_TieOnRecordedAt_IsDeterministic`
  - create two positions with same `recorded_at` but different `created_at` (or IDs)
  - decide expected behavior (requires code change to order by `recorded_at DESC, created_at DESC` or similar); add test once behavior is defined

---

### Issue: Add missing UI formatting tests for `FormatItemWithPosition` and timeline formatting
**Why**
`FormatItemWithPosition` and `FormatPositionForTimeline` are used in CLI output, but only `FormatPosition` and `FormatRelativeTime` are directly tested.

**Add tests**
- `TestFormatItemWithPosition_NoPosition`
  - `pos=nil` should include “no position” and item name
- `TestFormatItemWithPosition_WithLabel`
  - ensures label appears and coords do not (current behavior picks label)
- `TestFormatItemWithPosition_WithoutLabel_UsesCoords`
  - label nil/empty should include formatted coords
- `TestFormatPositionForTimeline_WithAndWithoutLabel`
  - assert it includes formatted date string and indentation prefix `"  "`

---

### Issue: Add missing CLI integration tests for invalid input handling (`position add`)
**Why**
The only integration test covers the happy path. The CLI has validation paths for required flags, coordinate ranges, and timestamp parsing.

**Add integration tests in `test/integration_test.go`**
- `TestAdd_MissingRequiredFlags`
  - `position add harper` (no flags) should fail and mention required flags (cobra error)
- `TestAdd_InvalidLatRange`
  - lat=100 should exit non-zero and print “latitude must be between -90 and 90”
- `TestAdd_InvalidLngRange`
  - lng=200 should exit non-zero and print “longitude must be between -180 and 180”
- `TestAdd_InvalidTimestamp`
  - `--at not-a-time` should exit non-zero and print “invalid timestamp format”
- `TestAdd_EmptyLabelTreatedAsNoLabel`
  - `--label ""` should behave like no label (validate output doesn’t contain double spaces / label formatting)

---

### Issue: Add missing CLI integration tests for empty states and error states (`list/current/timeline/remove`)
**Why**
Code has specific messaging for empty list and for items with no positions; these aren’t tested.

**Add integration tests**
- `TestList_WhenNoItems_PrintsHint`
  - new DB, `position list` should output: “No items tracked yet…”
- `TestCurrent_ItemNotFound`
  - `position current missing` should exit non-zero and include `item 'missing' not found`
- `TestTimeline_ItemNotFound`
  - `position timeline missing` should exit non-zero and include `item 'missing' not found`
- `TestTimeline_NoPositions_PrintsNoHistory`
  - create item with no positions (may require a helper command or direct DB insertion in test; simplest: call `add`? but that creates positions—so insert item directly using `internal/db` in a test that runs the binary, or extend CLI; if staying black-box, this isn’t currently possible without internal access)
- `TestRemove_ItemNotFound`
  - `position remove missing --confirm` should exit non-zero and include `item 'missing' not found`

---

### Issue: Add unit tests for MCP server creation and tool handlers (`internal/mcp`)
**Why**
MCP logic is non-trivial and currently has no tests. At minimum, validate error handling and DB side-effects of handlers.

**Add tests**
- `TestNewServer_NilDBReturnsError` (unit test; deterministic)
- Handler tests using a temp sqlite DB (via `db.InitDB`) and calling handler methods directly:
  - `TestHandleAddPosition_CreatesItemAndPosition`
  - `TestHandleAddPosition_RejectsInvalidCoordinates`
  - `TestHandleAddPosition_InvalidAtReturnsError`
  - `TestHandleGetCurrent_ItemNotFound`
  - `TestHandleGetCurrent_NoPositionReturnsError`
  - `TestHandleGetTimeline_ReturnsPositionsNewestFirst`
  - `TestHandleRemoveItem_RemovesAndCascadesPositions`
  - `TestHandleListItems_IncludesCurrentPositionWhenExists`
  - `TestHandleListItems_CurrentPositionOmittedWhenNone`
- Resource test:
  - `TestHandleItemsResource_ReturnsJSONWithCount`
  - parse returned JSON and assert `count` matches inserted items

(These can avoid spinning up a real MCP transport; just instantiate `Server{db: ..., mcp: ...}` via `NewServer` and call methods.)

---

### Issue: Strengthen `GetDefaultDBPath` tests for XDG behavior
**Why**
Current test only asserts “absolute” and contains “position” using a custom `contains` helper. It doesn’t verify `XDG_DATA_HOME` override behavior.

**Add tests**
- `TestGetDefaultDBPath_UsesXDG_DATA_HOMEWhenSet`
  - set `XDG_DATA_HOME` to temp dir, call `GetDefaultDBPath`, assert it equals `<xdg>/position/position.db`
  - restore env var in cleanup
- `TestGetDefaultDBPath_WhenXDGNotSet_UsesHomeLocalShare`
  - hard to assert exact home in CI, but can at least assert suffix `/.local/share/position/position.db` when `XDG_DATA_HOME` empty (may require skipping if `os.UserHomeDir` fails; or assert it ends with `position/position.db` and contains `.local/share`)

---

### Issue: Add DB tests for scanning/parsing failures (UUID parsing)
**Why**
`scanItem/scanPosition` now return parse errors if stored UUID strings are invalid. There is no test to ensure this behavior works (and continues to work).

**Add tests**
- `TestGetItemByName_InvalidUUIDInRow_ReturnsError`
  - insert a row into `items` manually with `id='not-a-uuid'`, then call `GetItemByName` and assert error contains “failed to parse item ID”
- `TestGetCurrentPosition_InvalidUUIDInRow_ReturnsError`
  - manually insert a position with invalid `id` or `item_id` and assert parse error

---
