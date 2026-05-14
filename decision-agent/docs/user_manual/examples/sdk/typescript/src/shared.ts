import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline/promises';
import { fileURLToPath } from 'node:url';

import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

const examplesRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');
const envFile = path.join(examplesRoot, '.env');
const stateDir = path.join(examplesRoot, '.tmp');
export const stateFile = path.join(stateDir, 'state.env');

function parseKvFile(file: string): Record<string, string> {
  if (!fs.existsSync(file)) return {};

  const values: Record<string, string> = {};
  for (const rawLine of fs.readFileSync(file, 'utf8').split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) continue;
    const match = line.match(/^([A-Za-z_][A-Za-z0-9_]*)=(.*)$/);
    if (!match) continue;

    let value = match[2].replace(/\s+#.*$/, '').trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    values[match[1]] = value;
  }
  return values;
}

function loadKvFile(file: string): void {
  for (const [key, value] of Object.entries(parseKvFile(file))) {
    if (!process.env[key]) process.env[key] = value;
  }
}

function loadExampleEnv(): void {
  loadKvFile(envFile);
  loadKvFile(stateFile);

  process.env.KWEAVER_BASE_URL ||= 'http://127.0.0.1:13020';
  process.env.KWEAVER_NO_AUTH ||= '1';
  process.env.KWEAVER_BUSINESS_DOMAIN ||= 'bd_public';
  process.env.KWEAVER_SDK_PACKAGE_SOURCE ||= 'remote';
  process.env.KWEAVER_SDK_REMOTE_PACKAGE ||= '@kweaver-ai/kweaver-sdk';
  process.env.KWEAVER_SDK_TS_DIR ||= '../../../../../kweaver-sdk/packages/typescript';
  process.env.KWEAVER_LLM_ID ||= 'xxx';
  process.env.KWEAVER_LLM_NAME ||= 'deepseek-v3';
  process.env.AGENT_VERSION ||= 'v0';
}

function writeState(values: Record<string, string>): void {
  fs.mkdirSync(stateDir, { recursive: true });
  const current = parseKvFile(stateFile);
  const next = { ...current, ...values };
  const content = Object.entries(next)
    .filter(([, value]) => value !== '')
    .map(([key, value]) => `${key}=${value}`)
    .join('\n');
  fs.writeFileSync(stateFile, `${content}\n`);
  Object.assign(process.env, values);
}

loadExampleEnv();

export function configureKweaver(): void {
  const baseUrl = process.env.KWEAVER_BASE_URL ?? 'http://127.0.0.1:13020';
  const noAuth = (process.env.KWEAVER_NO_AUTH ?? '1').toLowerCase();
  const auth = ['1', 'true', 'yes'].includes(noAuth) ? false : undefined;

  kweaver.configure({
    baseUrl,
    auth,
    accessToken: process.env.KWEAVER_TOKEN,
    businessDomain: process.env.KWEAVER_BUSINESS_DOMAIN ?? 'bd_public',
  });
}

export function stateSet(values: Record<string, string | undefined>): void {
  const normalized: Record<string, string> = {};
  for (const [key, value] of Object.entries(values)) {
    if (value) normalized[key] = value;
  }
  if (Object.keys(normalized).length > 0) writeState(normalized);
}

export function stateClear(keys: string[]): void {
  fs.mkdirSync(stateDir, { recursive: true });
  const current = parseKvFile(stateFile);
  for (const key of keys) {
    delete current[key];
    delete process.env[key];
  }
  const content = Object.entries(current).map(([key, value]) => `${key}=${value}`).join('\n');
  fs.writeFileSync(stateFile, content ? `${content}\n` : '');
}

export async function createExampleAgent(): Promise<string> {
  configureKweaver();
  const created = await kweaver.getClient().agents.create({
    name: `example_sdk_agent_${new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14)}`,
    profile: 'Created by docs/user_manual/examples/sdk/typescript',
    product_key: 'DIP',
    config: buildMinimalConfig(),
  });

  stateSet({
    AGENT_ID: created.id,
    AGENT_VERSION: created.version ?? 'v0',
  });
  return created.id;
}

export async function ensureAgentId(options: { createIfMissing?: boolean } = {}): Promise<string> {
  if (process.env.AGENT_ID) return process.env.AGENT_ID;

  if (options.createIfMissing && process.stdin.isTTY && process.stdout.isTTY) {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    const answer = await rl.question('AGENT_ID is missing. Create a temporary Agent now? [y/N] ');
    rl.close();
    if (/^(y|yes)$/i.test(answer.trim())) {
      return createExampleAgent();
    }
    throw new Error('Canceled: AGENT_ID is required.');
  }

  throw new Error(`AGENT_ID is required. Set AGENT_ID, add it to ${envFile}, or run the create target first.`);
}

export function buildMinimalConfig() {
  const llmId = process.env.KWEAVER_LLM_ID;
  const llmName = process.env.KWEAVER_LLM_NAME;

  if (!llmId || !llmName) {
    throw new Error('Please set KWEAVER_LLM_ID and KWEAVER_LLM_NAME before running flow.');
  }

  return {
    input: { fields: [{ name: 'query', type: 'string' }] },
    output: { default_format: 'markdown', variables: { answer_var: 'answer' } },
    llms: [
      {
        is_default: true,
        llm_config: {
          id: llmId,
          name: llmName,
          model_type: 'llm',
          temperature: 1,
          top_p: 1,
          top_k: 1,
          frequency_penalty: 0,
          presence_penalty: 0,
          max_tokens: 1000,
          retrieval_max_tokens: 32,
        },
      },
    ],
    memory: { is_enabled: false },
    related_question: { is_enabled: false },
    plan_mode: { is_enabled: false },
  };
}

export { kweaver };
