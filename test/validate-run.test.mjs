import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';

import { runCli, validateRun } from '../scripts/validate-run.mjs';

async function makeRun(overrides = {}) {
  const runDir = await fs.mkdtemp(path.join(os.tmpdir(), 'ai-auto-test-'));
  const runId = '20260820-100000-test';
  const status =
    overrides.status ??
    [
      'run_id,case_id,browser_status,manual_required,last_execution_id,updated_at',
      `${runId},TC-01,passed,false,${runId}-TC-01-a1,2026-08-20T10:01:00+08:00`,
    ].join('\n');
  const record = {
    executionId: `${runId}-TC-01-a1`,
    runId,
    caseId: 'TC-01',
    attempt: 1,
    status: 'passed',
    startedAt: '2026-08-20T10:00:00+08:00',
    finishedAt: '2026-08-20T10:01:00+08:00',
    steps: [
      {
        stepIndex: 1,
        action: '打开页面',
        expected: '显示首页',
        observed: '页面显示首页',
        result: 'passed',
      },
    ],
    finalUrl: 'http://localhost/example',
    pageTitle: '示例',
    evidence: [{ kind: 'screenshot', uri: 'tc-01.png' }],
  };
  const records = overrides.records ?? [record];
  const summary =
    overrides.summary ??
    [
      `# ${runId}`,
      '| 状态 | 数量 |',
      '| --- | --- |',
      '| passed | 1 |',
      '| failed | 0 |',
      '| blocked | 0 |',
      '| inconclusive | 0 |',
    ].join('\n');

  await fs.writeFile(path.join(runDir, 'case-status.csv'), `\uFEFF${status}\n`, 'utf8');
  await fs.writeFile(
    path.join(runDir, 'case-executions.jsonl'),
    `${records.map((item) => JSON.stringify(item)).join('\n')}\n`,
    'utf8',
  );
  await fs.writeFile(path.join(runDir, 'summary.md'), `${summary}\n`, 'utf8');
  await fs.writeFile(path.join(runDir, 'tc-01.png'), 'not-empty', 'utf8');
  return { runDir, runId, record };
}

test('accepts a consistent run', async (context) => {
  const fixture = await makeRun();
  context.after(() => fs.rm(fixture.runDir, { recursive: true, force: true }));

  const report = await validateRun(fixture.runDir, { strict: true });
  assert.equal(report.valid, true);
  assert.equal(report.stats.cases, 1);
  assert.equal(report.stats.executions, 1);
});

test('rejects duplicate execution IDs and attempts', async (context) => {
  const base = await makeRun();
  context.after(() => fs.rm(base.runDir, { recursive: true, force: true }));
  await fs.writeFile(
    path.join(base.runDir, 'case-executions.jsonl'),
    `${JSON.stringify(base.record)}\n${JSON.stringify(base.record)}\n`,
    'utf8',
  );

  const report = await validateRun(base.runDir);
  assert.equal(report.valid, false);
  assert.ok(report.errors.some((item) => item.code === 'DUPLICATE_EXECUTION_ID'));
  assert.ok(report.errors.some((item) => item.code === 'DUPLICATE_ATTEMPT'));
});

test('rejects an abnormal status without manual follow-up', async (context) => {
  const runId = '20260820-100000-test';
  const fixture = await makeRun({
    status: [
      'run_id,case_id,browser_status,manual_required,last_execution_id,updated_at',
      `${runId},TC-01,failed,false,${runId}-TC-01-a1,2026-08-20T10:01:00+08:00`,
    ].join('\n'),
    records: [
      {
        executionId: `${runId}-TC-01-a1`,
        runId,
        caseId: 'TC-01',
        attempt: 1,
        status: 'failed',
        startedAt: '2026-08-20T10:00:00+08:00',
        finishedAt: '2026-08-20T10:01:00+08:00',
        steps: [
          {
            stepIndex: 1,
            action: '打开页面',
            expected: '显示不存在的内容',
            observed: '未显示',
            result: 'failed',
          },
        ],
        evidence: [{ kind: 'screenshot', uri: 'tc-01.png' }],
      },
    ],
    summary: [
      `# ${runId}`,
      '| 状态 | 数量 |',
      '| --- | --- |',
      '| passed | 0 |',
      '| failed | 1 |',
      '| blocked | 0 |',
      '| inconclusive | 0 |',
    ].join('\n'),
  });
  context.after(() => fs.rm(fixture.runDir, { recursive: true, force: true }));

  const report = await validateRun(fixture.runDir);
  assert.equal(report.valid, false);
  assert.ok(report.errors.some((item) => item.code === 'MANUAL_REQUIRED_MISMATCH'));
});

test('rejects a stale status projection and mismatched summary counts', async (context) => {
  const runId = '20260820-100000-test';
  const fixture = await makeRun({
    status: [
      'run_id,case_id,browser_status,manual_required,last_execution_id,updated_at',
      `${runId},TC-01,passed,false,missing-execution,2026-08-20T10:01:00+08:00`,
    ].join('\n'),
    summary: [
      `# ${runId}`,
      '| 状态 | 数量 |',
      '| --- | --- |',
      '| passed | 0 |',
      '| failed | 1 |',
      '| blocked | 0 |',
      '| inconclusive | 0 |',
    ].join('\n'),
  });
  context.after(() => fs.rm(fixture.runDir, { recursive: true, force: true }));

  const report = await validateRun(fixture.runDir);
  assert.equal(report.valid, false);
  assert.ok(report.errors.some((item) => item.code === 'STALE_LAST_EXECUTION'));
  assert.ok(report.errors.some((item) => item.code === 'SUMMARY_COUNT_MISMATCH'));
});

test('rejects credentials and evidence paths outside the run directory', async (context) => {
  const fixture = await makeRun();
  context.after(() => fs.rm(fixture.runDir, { recursive: true, force: true }));

  const unsafeRecord = {
    ...fixture.record,
    evidence: [{ kind: 'screenshot', uri: '../outside.png' }],
  };
  await fs.writeFile(
    path.join(fixture.runDir, 'case-executions.jsonl'),
    `${JSON.stringify(unsafeRecord)}\n`,
    'utf8',
  );
  await fs.appendFile(path.join(fixture.runDir, 'summary.md'), '\n密码：example-secret-value\n', 'utf8');

  const report = await validateRun(fixture.runDir);
  assert.equal(report.valid, false);
  assert.ok(report.errors.some((item) => item.code === 'EVIDENCE_PATH_ESCAPE'));
  assert.ok(report.errors.some((item) => item.code === 'POSSIBLE_CREDENTIAL_LEAK'));
});

test('CLI returns success for a valid strict run', async (context) => {
  const fixture = await makeRun();
  context.after(() => fs.rm(fixture.runDir, { recursive: true, force: true }));
  const stdout = [];
  const stderr = [];

  const exitCode = await runCli([fixture.runDir, '--strict'], {
    log: (message) => stdout.push(message),
    error: (message) => stderr.push(message),
  });

  assert.equal(exitCode, 0, stderr.join('\n'));
  assert.match(stdout.join('\n'), /VALIDATION PASSED/);
  assert.match(stdout.join('\n'), /cases: 1/);
});
