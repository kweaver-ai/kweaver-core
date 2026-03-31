/**
 * 整合国际化资源
 */
import global_zh from './global/zh-CN.json';
import graphList_zh from './graphList/zh-CN.json';
import knowledgeNetwork_zh from './knowledgeNetwork/zh-CN.json';
import cognitiveService_zh from './cognitiveService/zh-CN.json';
import prompt_zh from './prompt/zh-CN.json';
import benchmarkTask_zh from './benckmarkTask/zh-CN.json';
import agent_zh from './agent/zh-CN.json';
import agentTool_zh from './agentTool/zh-CN.json';
import agentCommonConfig_zh from './agentCommonConfig/zh-CN.json';
import federalAi_zh from './federalAi/zh-CN.json';
import docLib_zh from './doc-lib/zh-CN.json';
import time_zh from './time/zh-CN.json';
import error_zh from './error/zh-CN.json';
import dataAgent_zh from './data-agent/zh-CN.json';
import superAssistant_zh from './super-assistant/zh-CN.json';
import dipChat_zh from './dip-chat/zh-CN.json';
import businessDomain_zh from './business-domian/zh-CN.json';

import global_tw from './global/zh-TW.json';
import graphList_tw from './graphList/zh-TW.json';
import knowledgeNetwork_tw from './knowledgeNetwork/zh-TW.json';
import cognitiveService_tw from './cognitiveService/zh-TW.json';
import prompt_tw from './prompt/zh-TW.json';
import benchmarkTask_tw from './benckmarkTask/zh-TW.json';
import agent_tw from './agent/zh-TW.json';
import agentTool_tw from './agentTool/zh-TW.json';
import agentCommonConfig_tw from './agentCommonConfig/zh-TW.json';
import federalAi_tw from './federalAi/zh-TW.json';
import docLib_tw from './doc-lib/zh-TW.json';
import time_tw from './time/zh-TW.json';
import error_tw from './error/zh-TW.json';
import dataAgent_tw from './data-agent/zh-TW.json';
import superAssistant_tw from './super-assistant/zh-TW.json';
import dipChat_tw from './dip-chat/zh-TW.json';
import businessDomain_tw from './business-domian/zh-TW.json';

import global_en from './global/en-US.json';
import graphList_en from './graphList/en-US.json';
import knowledgeNetwork_en from './knowledgeNetwork/en-US.json';
import cognitiveService_en from './cognitiveService/en-US.json';
import prompt_en from './prompt/en-US.json';
import benchmarkTask_en from './benckmarkTask/en-US.json';
import agent_en from './agent/en-US.json';
import agentTool_en from './agentTool/en-US.json';
import agentCommonConfig_en from './agentCommonConfig/en-US.json';
import federalAi_en from './federalAi/en-US.json';
import docLib_en from './doc-lib/en-US.json';
import time_en from './time/en-US.json';
import error_en from './error/en-US.json';
import dataAgent_en from './data-agent/en-US.json';
import superAssistant_en from './super-assistant/en-US.json';
import dipChat_en from './dip-chat/en-US.json';
import businessDomain_en from './business-domian/en-US.json';

const zh_CN = {
  ...global_zh,
  ...graphList_zh,
  ...knowledgeNetwork_zh,
  ...cognitiveService_zh,
  ...prompt_zh,
  ...benchmarkTask_zh,
  ...agent_zh,
  ...agentTool_zh,
  ...agentCommonConfig_zh,
  ...federalAi_zh,
  ...docLib_zh,
  ...time_zh,
  ...error_zh,
  ...dataAgent_zh,
  ...superAssistant_zh,
  ...dipChat_zh,
  ...businessDomain_zh,
};

const zh_TW = {
  ...global_tw,
  ...graphList_tw,
  ...knowledgeNetwork_tw,
  ...cognitiveService_tw,
  ...prompt_tw,
  ...benchmarkTask_tw,
  ...agent_tw,
  ...agentTool_tw,
  ...agentCommonConfig_tw,
  ...federalAi_tw,
  ...docLib_tw,
  ...time_tw,
  ...error_tw,
  ...dataAgent_tw,
  ...superAssistant_tw,
  ...dipChat_tw,
  ...businessDomain_tw,
};

const en_US = {
  ...global_en,
  ...graphList_en,
  ...knowledgeNetwork_en,
  ...cognitiveService_en,
  ...prompt_en,
  ...benchmarkTask_en,
  ...agent_en,
  ...agentTool_en,
  ...agentCommonConfig_en,
  ...federalAi_en,
  ...docLib_en,
  ...time_en,
  ...error_en,
  ...dataAgent_en,
  ...superAssistant_en,
  ...dipChat_en,
  ...businessDomain_en,
};

const locales = {
  'zh-CN': zh_CN,
  'zh-TW': zh_TW,
  'en-US': en_US,
};

export default locales;
