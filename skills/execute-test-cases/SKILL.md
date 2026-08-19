---
name: execute-test-cases
description: Execute existing positive-path web test cases with Cursor Agent Window's native Browser, record evidence and verdicts, and produce a human follow-up list. Use for a Cursor-first AI browser-testing proof of concept when the user provides test cases, a test URL, and dedicated test accounts. Do not use to modify application code or test production systems.
---

# Execute Test Cases

Use Cursor Agent Window's native Browser through `/use-browser`. Do not install Playwright, Browser Use, or another browser tool for this proof of concept.

## Keep the scope narrow

- Test only through the visible browser UI.
- Treat source test cases as read-only.
- Do not read or modify application code, call internal APIs, or access databases.
- Do not fix failures while testing.
- Use only a test environment and dedicated test accounts.
- Execute every input case sequentially. Assume cases are independent.

## Require inputs

Obtain these before execution:

1. Test environment URL.
2. A test-case file or pasted case list containing case ID, title, account, preconditions, ordered steps, and an expected result for every step.
3. The required named test accounts and their login methods.

Accept normal username/password login. For SMS mock login marked `浏览器自动填入`, enter the phone number, click to request the code, and wait for the page to fill the code field. If it does not fill before timeout, mark the case blocked.

Before creating a run, use `/use-browser` to open the target URL and confirm that visible page content can be read. If the Browser tool is unavailable, record the original error and mark the run `blocked`. Do not silently substitute another tool and do not mark business cases as failed.

## Prepare the run

Create `.ai-auto-test/results/<run-id>/summary.md` in the target project. Use a timestamp-based run ID. Record the URL, start time, Cursor version when available, case count, and account aliases. Never write passwords or SMS codes to result files.

Initialize every case as `pending` (`待测试`).

## Execute one case at a time

For each case:

1. Select the named account. Reuse the current authenticated session when the account matches; otherwise log out and log in with the required account.
2. Establish the stated preconditions only through the browser. If they cannot be established, mark the case `blocked`.
3. Follow the business steps in order. Preserve the business intent and expected results.
4. Adapt only UI mechanics when necessary, such as closing a harmless popup, scrolling, waiting, or switching tabs. Record any such deviation.
5. After every step, verify its expected result from the rendered page. Do not infer success only from a click completing.
6. Capture a Browser screenshot on final success and immediately on failure, blocking, or uncertainty.
7. Append the result to `summary.md` before proceeding, so an interrupted run can resume from the first pending case.

Use simple unique values for clearly open-ended positive-path data such as a new customer name, and record the actual value. Do not replace explicit fixed values or change amounts, roles, products, states, or expected results.

## Assign exactly one verdict

- `passed` (`已通过`): every step ran and every expected result was directly observed.
- `failed` (`不通过`): execution completed far enough to observe behavior that contradicts an expected result.
- `blocked` (`测试受阻`): the case could not be executed because of login, permissions, missing data, environment, browser capability, or another prerequisite.
- `inconclusive` (`无法判断`): execution occurred but the visible UI did not provide enough evidence to decide.

Never convert blocked or inconclusive into passed. Record the failed or blocked step, concise reason, visible evidence, screenshot capture, and any relevant console or network error available from Cursor Browser.

## Finish the run

At the top of `summary.md`, write totals for all verdicts. Add an `人工处理清单` containing every failed, blocked, and inconclusive case. Passed cases require no manual action in this proof of concept.

Report whether the proof of concept met these checks:

- Most selected positive-path cases completed without human intervention.
- Every case has a verdict and screenshot evidence.
- Every abnormal verdict identifies a step and understandable reason.
- The human follow-up list is complete.
