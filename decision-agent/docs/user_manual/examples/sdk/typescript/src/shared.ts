import kweaver from '@kweaver-ai/kweaver-sdk/kweaver';

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
