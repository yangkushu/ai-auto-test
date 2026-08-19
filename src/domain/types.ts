/** Shared contract for the Browser-only MVP. */
export const browserStatuses = [
  'pending',
  'passed',
  'failed',
  'blocked',
  'inconclusive',
  'retest_pending',
] as const;
export type BrowserStatus = (typeof browserStatuses)[number];

export type FailureCategory =
  | 'product_bug'
  | 'test_data'
  | 'environment'
  | 'browser_capability'
  | 'permission'
  | 'unknown';

export interface CaseStatus {
  runId: string;
  caseId: string;
  browserStatus: BrowserStatus;
  manualRequired: boolean;
  lastExecutionId?: string;
  updatedAt: string;
}

export interface ExecutionEvidence {
  kind: 'screenshot' | 'network_log' | 'console_log';
  uri: string;
}

export interface StepExecution {
  stepIndex: number;
  action: string;
  expected: string;
  observed: string;
  result: 'passed' | 'failed' | 'blocked' | 'inconclusive';
}

export interface CaseExecution {
  executionId: string;
  runId: string;
  caseId: string;
  attempt: number;
  status: 'passed' | 'failed' | 'blocked' | 'inconclusive';
  startedAt: string;
  finishedAt: string;
  steps: StepExecution[];
  reason?: string;
  failureCategory?: FailureCategory;
  finalUrl?: string;
  pageTitle?: string;
  evidence: ExecutionEvidence[];
}
