---
name: execute-test-cases
description: Execute existing positive-path web test cases with Cursor Agent Window's native Browser, persist auditable status, execution history, screenshots, and human follow-up items, and resume or retest without overwriting history. Use for Cursor-first browser testing when the user provides a test URL, read-only test cases, and dedicated test accounts. Do not use to modify application code or test production systems.
---

# Execute Test Cases

Use Cursor Agent Window's native Browser. Prefer the verified `/use-browser` entry. Do not install or silently substitute another browser tool.

## Keep scope narrow

- Test only through the visible browser UI.
- Treat source cases as read-only.
- Do not read or modify application code, call internal APIs, access databases, or fix failures.
- Use a test environment and dedicated accounts.
- Execute every input case sequentially and treat cases as independent.
- Never write passwords or verification codes to results.

## Require inputs

Obtain the environment URL, a case file containing IDs, accounts, preconditions, ordered steps and per-step expectations, and named test accounts with login methods.

For SMS mock login marked `浏览器自动填入`, enter the phone number, request the code, and wait for the page to fill it. Mark the case `blocked` if it does not fill before timeout.

## Run Browser Preflight

Open the target URL and confirm visible content can be read before executing cases. If Browser is unavailable, record the original error and mark the run `blocked`; do not mark business cases as failed.

## Persist the run

Create `.ai-auto-test/results/<run-id>/` with:

- `case-status.csv`: current status projection;
- `case-executions.jsonl`: append-only attempt history;
- `summary.md`: run metadata, counts and human follow-up list;
- screenshot evidence files.

Use UTF-8 without BOM for JSONL and Markdown, and UTF-8 BOM for CSV. Initialize every case as `pending` (`待测试`). Limit any extra filesystem permission to the current results directory.

## Execute one case at a time

For each `pending` or `retest_pending` case:

1. Select the named account; reuse the session only when the case permits it.
2. Establish preconditions only through the browser; otherwise return `blocked`.
3. Follow steps in order and preserve business intent.
4. Adapt only UI mechanics such as harmless popups, scrolling, waiting, or tab switching; record deviations.
5. Verify every expectation from rendered content. A successful click is not sufficient evidence.
6. Capture the final or abnormal page and persist the screenshot.
7. Append a complete execution record before updating the status row.
8. Update `last_execution_id`, status, `manual_required`, and time before continuing.

Use simple unique values only for explicitly open-ended positive-path data. Never change fixed amounts, roles, products, states, or expected results.

## Assign one verdict

- `passed` (`已通过`): every expectation was directly observed.
- `failed` (`不通过`): visible behavior contradicted an expectation.
- `blocked` (`测试受阻`): login, permission, environment, data, Browser, or another prerequisite prevented execution.
- `inconclusive` (`无法判断`): execution occurred, but visible evidence was insufficient.

Set `manual_required=false` only for `passed`; set it to `true` for every abnormal verdict. Never convert blocked or inconclusive into passed.

## Resume and retest

Reuse the original run ID. Read status and history first, execute only `pending`/`retest_pending`, and never delete or change an existing JSONL line. Allocate the next attempt number and a unique execution ID. If status and history disagree after interruption, treat append-only execution history as the source for rebuilding status and record the repair.

## Finish and validate

Write current verdict counts, every case's latest result, evidence references, screenshot hashes when available, and all abnormal cases in `人工处理清单`.

When `scripts/validate-run.mjs` exists in the workspace, run:

```text
node scripts/validate-run.mjs .ai-auto-test/results/<run-id> --strict
```

Do not report the run complete if validation errors remain. Report warnings and any evidence that could not be persisted.
