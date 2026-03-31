import { Button, Drawer, Form, Input } from 'antd';
import { useDipChatStore } from '@/components/DipChat/store';
import { getAgentInputDisplayFields } from '@/components/DipChat/utils';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import './index.less';
import React, { useEffect } from 'react';
import _ from 'lodash';
import { useDeepCompareEffect } from '@/hooks';
import DipIcon from '@/components/DipIcon';
import intl from 'react-intl-universal';
const FormItem = Form.Item;
const AgentInputParamDrawer = () => {
  const {
    dipChatStore: { agentDetails, showAgentInputParamsDrawer },
    setDipChatStore,
    getDipChatStore,
  } = useDipChatStore();
  const [form] = Form.useForm();
  const filelds = getAgentInputDisplayFields(agentDetails?.config);

  useEffect(() => {
    const { agentInputParamsFormValue } = getDipChatStore();
    if (!_.isEmpty(agentInputParamsFormValue)) {
      form.setFieldsValue(agentInputParamsFormValue);
    } else {
      form.resetFields();
    }
  }, []);

  useDeepCompareEffect(() => {
    if (filelds.length > 0) {
      setDipChatStore({ agentInputParamForm: form });
    } else {
      setDipChatStore({ agentInputParamForm: null });
    }
  }, [filelds]);

  return (
    filelds.length > 0 && (
      <Drawer
        forceRender
        rootClassName="AgentInputParamDrawer"
        title={intl.get('dipChat.moreParams')}
        closable={false}
        open={showAgentInputParamsDrawer}
        onClose={() => {
          setDipChatStore({ showAgentInputParamsDrawer: false });
        }}
        getContainer={false}
        extra={
          <Button
            onClick={() => setDipChatStore({ showAgentInputParamsDrawer: false })}
            type="text"
            icon={<DipIcon type="icon-dip-close" />}
          />
        }
      >
        <Form
          layout="vertical"
          form={form}
          onValuesChange={(changedValues, values) => {
            setDipChatStore({ agentInputParamsFormValue: values, agentInputParamsFormErrorFields: [] });
          }}
        >
          {filelds.map((item: any) => {
            if (item.type === 'object') {
              return (
                <FormItem
                  key={item.name}
                  label={item.name}
                  name={item.name}
                  // rules={[{ required: true, message: `${item.name}不能为空` }]}
                >
                  <AdMonacoEditor
                    placeholder={intl.get('dipChat.enterParamName', { name: item.name })}
                    bordered
                    height="auto"
                    minHeight={66}
                    maxHeight={260}
                    options={{
                      minimap: { enabled: false },
                      fontSize: 14,
                      lineHeight: 22,
                      tabSize: 2,
                      insertSpaces: true,
                      readOnly: false,
                      scrollbar: {
                        alwaysConsumeMouseWheel: false, // 禁用Monaco的默认滚轮事件
                      },
                      lineNumbersMinChars: 4,
                      unicodeHighlight: {
                        ambiguousCharacters: false, // 关闭中文符号高亮报警
                      },
                      // 关闭智能提示
                      suggestOnTriggerCharacters: false, // 禁用触发字符的建议
                      quickSuggestions: false, // 禁用快速建议
                      parameterHints: { enabled: false }, // 禁用参数提示
                      hover: { enabled: false }, // 禁用悬停提示
                      wordBasedSuggestions: false, // 禁用基于单词的建议
                      snippetSuggestions: 'none', // 禁用代码片段建议
                    }}
                  />
                </FormItem>
              );
            }
            return (
              <FormItem
                key={item.name}
                label={item.name}
                name={item.name}
                // rules={[{ required: true, message: `${item.name}不能为空` }]}
              >
                <Input placeholder={intl.get('dipChat.enterParamName', { name: item.name })} />
              </FormItem>
            );
          })}
        </Form>
      </Drawer>
    )
  );
};

export default AgentInputParamDrawer;
