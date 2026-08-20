#!/usr/bin/env node

import fs from 'node:fs/promises';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const STATUS_VALUES = new Set([
  'pending',
  'passed',
  'failed',
  'blocked',
  'inconclusive',
  'retest_pending',
]);
const EXECUTION_VALUES = new Set(['passed', 'failed', 'blocked', 'inconclusive']);
const REQUIRED_STATUS_HEADERS = [
  'run_id',
  'case_id',
  'browser_status',
  'manual_required',
  'last_execution_id',
  'updated_at',
];

function parseCsv(text) {
  const source = text.replace(/^\uFEFF/, '');
  const rows = [];
  let row = [];
  let field = '';
  let quoted = false;

  for (let index = 0; index < source.length; index += 1) {
    const char = source[index];
    if (quoted) {
      if (char === '"' && source[index + 1] === '"') {
        field += '"';
        index += 1;
      } else if (char === '"') {
        quoted = false;
      } else {
        field += char;
      }
      continue;
    }

    if (char === '"') {
      quoted = true;
    } else if (char === ',') {
      row.push(field);
      field = '';
    } else if (char === '\n') {
      row.push(field.replace(/\r$/, ''));
      rows.push(row);
      row = [];
      field = '';
    } else {
      field += char;
    }
  }

  if (quoted) throw new Error('CSV contains an unterminated quoted field');
  if (field.length > 0 || row.length > 0) {
    row.push(field.replace(/\r$/, ''));
    rows.push(row);
  }
  return rows.filter((item) => item.some((value) => value !== ''));
}

function parseIsoTimestamp(value) {
  return typeof value === 'string' && value.length > 0 && !Number.isNaN(Date.parse(value));
}

function isInside(root, candidate) {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}

function issue(code, message) {
  return { code, message };
}

async function readRequiredFile(runDir, fileName, errors) {
  const filePath = path.join(runDir, fileName);
  try {
    return await fs.readFile(filePath, 'utf8');
  } catch (error) {
    errors.push(issue('MISSING_FILE', `${fileName}: ${error.message}`));
    return null;
  }
}

function findSummaryCount(summary, status) {
  const escaped = status.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = summary.match(new RegExp(`\\|\\s*${escaped}\\s*\\|\\s*(\\d+)\\s*\\|`, 'i'));
  return match ? Number(match[1]) : null;
}

export async function validateRun(inputDir, options = {}) {
  const runDir = path.resolve(inputDir);
  const strict = options.strict === true;
  const errors = [];
  const warnings = [];
  const statusText = await readRequiredFile(runDir, 'case-status.csv', errors);
  const executionText = await readRequiredFile(runDir, 'case-executions.jsonl', errors);
  const summaryText = await readRequiredFile(runDir, 'summary.md', errors);

  if (errors.length > 0) {
    return { valid: false, runDir, runId: null, errors, warnings, stats: {} };
  }

  let rows;
  try {
    rows = parseCsv(statusText);
  } catch (error) {
    errors.push(issue('INVALID_CSV', error.message));
    rows = [];
  }

  if (rows.length < 2) {
    errors.push(issue('EMPTY_STATUS', 'case-status.csv must contain a header and at least one case'));
  }

  const headers = rows[0] ?? [];
  for (const header of REQUIRED_STATUS_HEADERS) {
    if (!headers.includes(header)) {
      errors.push(issue('MISSING_STATUS_HEADER', `case-status.csv is missing ${header}`));
    }
  }

  const statusRows = rows.slice(1).map((values, rowIndex) => {
    if (values.length !== headers.length) {
      errors.push(
        issue(
          'STATUS_COLUMN_COUNT',
          `case-status.csv row ${rowIndex + 2} has ${values.length} columns; expected ${headers.length}`,
        ),
      );
    }
    return Object.fromEntries(headers.map((header, index) => [header, values[index] ?? '']));
  });

  const runIds = new Set(statusRows.map((row) => row.run_id).filter(Boolean));
  if (runIds.size !== 1) {
    errors.push(issue('STATUS_RUN_ID', 'case-status.csv must contain exactly one non-empty run_id'));
  }
  const runId = runIds.size === 1 ? [...runIds][0] : null;
  const statusesByCase = new Map();

  for (const row of statusRows) {
    if (!row.case_id) errors.push(issue('MISSING_CASE_ID', 'case-status.csv contains an empty case_id'));
    if (statusesByCase.has(row.case_id)) {
      errors.push(issue('DUPLICATE_CASE_ID', `duplicate status row for ${row.case_id}`));
    }
    statusesByCase.set(row.case_id, row);

    if (!STATUS_VALUES.has(row.browser_status)) {
      errors.push(issue('INVALID_STATUS', `${row.case_id} has invalid status ${row.browser_status}`));
    }
    if (!['true', 'false'].includes(row.manual_required)) {
      errors.push(
        issue('INVALID_MANUAL_REQUIRED', `${row.case_id} has invalid manual_required ${row.manual_required}`),
      );
    }
    if (!parseIsoTimestamp(row.updated_at)) {
      errors.push(issue('INVALID_UPDATED_AT', `${row.case_id} has invalid updated_at ${row.updated_at}`));
    }

    if (row.browser_status === 'passed' && row.manual_required !== 'false') {
      errors.push(issue('MANUAL_REQUIRED_MISMATCH', `${row.case_id} passed but manual_required is not false`));
    }
    if (['failed', 'blocked', 'inconclusive'].includes(row.browser_status) && row.manual_required !== 'true') {
      errors.push(
        issue('MANUAL_REQUIRED_MISMATCH', `${row.case_id} is abnormal but manual_required is not true`),
      );
    }
  }

  const records = [];
  const executionIds = new Set();
  const attemptsByCase = new Map();
  const executionLines = executionText.split(/\r?\n/).filter((line) => line.trim() !== '');

  for (let index = 0; index < executionLines.length; index += 1) {
    let record;
    try {
      record = JSON.parse(executionLines[index]);
    } catch (error) {
      errors.push(issue('INVALID_JSONL', `line ${index + 1}: ${error.message}`));
      continue;
    }
    records.push(record);

    if (!record.executionId || executionIds.has(record.executionId)) {
      errors.push(issue('DUPLICATE_EXECUTION_ID', `invalid or duplicate executionId on line ${index + 1}`));
    }
    executionIds.add(record.executionId);

    if (record.runId !== runId) {
      errors.push(issue('EXECUTION_RUN_ID', `${record.executionId} runId does not match case-status.csv`));
    }
    if (!statusesByCase.has(record.caseId)) {
      errors.push(issue('UNKNOWN_EXECUTION_CASE', `${record.executionId} references unknown case ${record.caseId}`));
    }
    if (!Number.isInteger(record.attempt) || record.attempt < 1) {
      errors.push(issue('INVALID_ATTEMPT', `${record.executionId} has invalid attempt ${record.attempt}`));
    }
    if (!EXECUTION_VALUES.has(record.status)) {
      errors.push(issue('INVALID_EXECUTION_STATUS', `${record.executionId} has invalid status ${record.status}`));
    }
    if (!parseIsoTimestamp(record.startedAt) || !parseIsoTimestamp(record.finishedAt)) {
      errors.push(issue('INVALID_EXECUTION_TIME', `${record.executionId} has invalid timestamps`));
    } else if (Date.parse(record.finishedAt) < Date.parse(record.startedAt)) {
      errors.push(issue('EXECUTION_TIME_ORDER', `${record.executionId} finishes before it starts`));
    }
    if (!Array.isArray(record.steps) || record.steps.length === 0) {
      errors.push(issue('MISSING_STEPS', `${record.executionId} has no step results`));
    } else {
      for (const [stepIndex, step] of record.steps.entries()) {
        if (!step.action || !step.expected || !step.observed || !EXECUTION_VALUES.has(step.result)) {
          errors.push(
            issue('INVALID_STEP', `${record.executionId} step ${stepIndex + 1} is missing required observations`),
          );
        }
      }
    }

    const caseAttempts = attemptsByCase.get(record.caseId) ?? [];
    if (caseAttempts.some((item) => item.attempt === record.attempt)) {
      errors.push(issue('DUPLICATE_ATTEMPT', `${record.caseId} has duplicate attempt ${record.attempt}`));
    }
    caseAttempts.push(record);
    attemptsByCase.set(record.caseId, caseAttempts);

    if (!Array.isArray(record.evidence) || record.evidence.length === 0) {
      errors.push(issue('MISSING_EVIDENCE', `${record.executionId} has no evidence`));
      continue;
    }

    for (const evidence of record.evidence) {
      if (!evidence?.kind || !evidence?.uri) {
        errors.push(issue('INVALID_EVIDENCE', `${record.executionId} contains invalid evidence`));
        continue;
      }
      if (evidence.kind !== 'screenshot') continue;
      if (path.isAbsolute(evidence.uri) || /^[a-z]:[\\/]/i.test(evidence.uri)) {
        errors.push(issue('EVIDENCE_PATH_ESCAPE', `${record.executionId} evidence must use a relative path`));
        continue;
      }
      if (/^[a-z][a-z0-9+.-]*:/i.test(evidence.uri)) {
        warnings.push(issue('REMOTE_EVIDENCE', `${record.executionId} screenshot is not locally persisted`));
        continue;
      }
      const evidencePath = path.resolve(runDir, evidence.uri);
      if (!isInside(runDir, evidencePath)) {
        errors.push(issue('EVIDENCE_PATH_ESCAPE', `${record.executionId} evidence escapes the run directory`));
        continue;
      }
      try {
        const stat = await fs.stat(evidencePath);
        if (!stat.isFile() || stat.size === 0) {
          errors.push(issue('EMPTY_EVIDENCE', `${record.executionId} evidence is empty`));
        }
      } catch {
        errors.push(issue('MISSING_EVIDENCE_FILE', `${record.executionId} is missing ${evidence.uri}`));
      }
    }
  }

  for (const [caseId, row] of statusesByCase.entries()) {
    const caseRecords = (attemptsByCase.get(caseId) ?? []).sort((left, right) => left.attempt - right.attempt);
    for (let index = 0; index < caseRecords.length; index += 1) {
      if (caseRecords[index].attempt !== index + 1) {
        errors.push(issue('ATTEMPT_GAP', `${caseId} attempts are not contiguous from 1`));
        break;
      }
    }

    if (['passed', 'failed', 'blocked', 'inconclusive'].includes(row.browser_status)) {
      if (!row.last_execution_id) {
        errors.push(issue('MISSING_LAST_EXECUTION', `${caseId} has a final status without last_execution_id`));
        continue;
      }
      const latest = caseRecords.at(-1);
      if (!latest || latest.executionId !== row.last_execution_id) {
        errors.push(issue('STALE_LAST_EXECUTION', `${caseId} does not point to its latest execution`));
      } else if (latest.status !== row.browser_status) {
        errors.push(issue('STATUS_EXECUTION_MISMATCH', `${caseId} status differs from its latest execution`));
      }
    } else if (row.browser_status === 'pending' && caseRecords.length > 0) {
      errors.push(issue('PENDING_WITH_HISTORY', `${caseId} is pending but already has execution history`));
    }
  }

  if (runId && !summaryText.includes(runId)) {
    errors.push(issue('SUMMARY_RUN_ID', 'summary.md does not contain the run ID'));
  }
  const currentCounts = Object.fromEntries([...STATUS_VALUES].map((status) => [status, 0]));
  for (const row of statusRows) {
    if (STATUS_VALUES.has(row.browser_status)) currentCounts[row.browser_status] += 1;
  }
  for (const status of ['passed', 'failed', 'blocked', 'inconclusive']) {
    const summaryCount = findSummaryCount(summaryText, status);
    if (summaryCount === null) {
      warnings.push(issue('SUMMARY_COUNT_MISSING', `summary.md does not expose a ${status} count`));
    } else if (summaryCount !== currentCounts[status]) {
      errors.push(
        issue(
          'SUMMARY_COUNT_MISMATCH',
          `summary.md reports ${status}=${summaryCount}; case-status.csv has ${currentCounts[status]}`,
        ),
      );
    }
  }

  const combinedOutput = `${statusText}\n${executionText}\n${summaryText}`;
  const credentialPatterns = [
    /["']?(?:password|passwd|pwd)["']?\s*[:=]\s*["']?[^\s,"'|}]+/i,
    /(?:密码|验证码)\s*[:：=]\s*[^\s,，|}]+/,
  ];
  if (credentialPatterns.some((pattern) => pattern.test(combinedOutput))) {
    errors.push(issue('POSSIBLE_CREDENTIAL_LEAK', 'result files may contain a password or verification code'));
  }

  const effectiveErrors = strict ? [...errors, ...warnings] : errors;
  return {
    valid: effectiveErrors.length === 0,
    runDir,
    runId,
    errors,
    warnings,
    stats: {
      cases: statusRows.length,
      executions: records.length,
      statuses: currentCounts,
    },
  };
}

function formatHuman(report) {
  const lines = [report.valid ? 'VALIDATION PASSED' : 'VALIDATION FAILED'];
  if (report.runId) lines.push(`run_id: ${report.runId}`);
  lines.push(`cases: ${report.stats.cases ?? 0}`);
  lines.push(`executions: ${report.stats.executions ?? 0}`);
  for (const item of report.errors) lines.push(`ERROR [${item.code}] ${item.message}`);
  for (const item of report.warnings) lines.push(`WARN  [${item.code}] ${item.message}`);
  return lines.join('\n');
}

export async function runCli(args, output = console) {
  const strict = args.includes('--strict');
  const json = args.includes('--json');
  const runDir = args.find((arg) => !arg.startsWith('--'));
  if (!runDir) {
    output.error('Usage: node scripts/validate-run.mjs <run-dir> [--strict] [--json]');
    return 2;
  }

  const report = await validateRun(runDir, { strict });
  output.log(json ? JSON.stringify(report, null, 2) : formatHuman(report));
  return report.valid ? 0 : 1;
}

const entryPath = process.argv[1] ? pathToFileURL(path.resolve(process.argv[1])).href : null;
if (entryPath === import.meta.url) {
  runCli(process.argv.slice(2)).then((exitCode) => {
    process.exitCode = exitCode;
  }).catch((error) => {
    console.error(error.stack ?? error.message);
    process.exitCode = 2;
  });
}
