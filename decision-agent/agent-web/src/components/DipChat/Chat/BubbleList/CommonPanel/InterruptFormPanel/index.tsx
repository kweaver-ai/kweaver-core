import { useDipChatStore } from '@/components/DipChat/store';
import { Alert, Button, Form, Input } from 'antd';
import { useDeepCompareMemo } from '@/hooks';
import _ from 'lodash';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import { SKILL_TYPE } from '@/components/DipChat/enum.ts';
import { isJSONString } from '@/utils/handle-function';

const FormItem = Form.Item;
type FieldType = {
  key: string;
  value: string;
  type: string;
};
const InterruptFormPanel = ({ chatItemIndex }: any) => {
  const {
    dipChatStore: { chatList },
    sendChat,
  } = useDipChatStore();
  const chatItem = chatList[chatItemIndex];
  const { interrupt, content } = chatItem;
  const progress = content.progress || [];
  const lastProgressItem = progress[progress.length - 1];
  const skillInfo = _.get(lastProgressItem, `skillInfo`) || {};
  const [form] = Form.useForm();

  const fields: FieldType[] = useDeepCompareMemo(() => {
    const tempArr: any = [];
    if (interrupt) {
      // const loop = (tool_args: any[]) => {
      //   // 先判断传进来的数组要不要执行深度递归
      //   // 如果元素的key: 'tool’并且 value.tool_args有值 则忽略其他元素，直接深度递归
      //   const targetArg = tool_args.find(
      //     (arg: any) => arg.key === 'tool' && Array.isArray(_.get(arg, 'value.tool_args'))
      //   );
      //   if (targetArg) {
      //     const args = _.get(targetArg, 'value.tool_args');
      //     if (args.length > 0) {
      //       loop(args);
      //     }
      //   } else {
      //     // 此时说明不需要深度递归，那么直接渲染即可
      //     tempArr = [...tempArr, ...tool_args];
      //   }
      // };
      // loop(interrupt.data.tool_args);

      return interrupt.data?.tool_args;
    }
    return tempArr;
  }, [interrupt]);

  const renderInput = (item: FieldType) => {
    if (item.type === 'string') {
      return <Input.TextArea autoSize={{ minRows: 1, maxRows: 5 }} placeholder={`请输入${item.key}`} />;
    }
    if (item.type === 'object' || item.type === 'dict' || item.type === 'list') {
      return (
        <AdMonacoEditor
          placeholder={`请输入${item.key}`}
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
      );
    }
    return <Input.TextArea autoSize={{ minRows: 1, maxRows: 5 }} placeholder={`请输入${item.key}`} />;
  };

  const updateInterrupt = () => {
    const formValues = form.getFieldsValue();
    // 只收集用户修改过的参数，仅包含 key 和 value
    const args = fields
      .filter(field => {
        let fieldValue = formValues[field.key];
        if (field.type === 'object' || field.type === 'dict' || field.type === 'list') {
          fieldValue = isJSONString(fieldValue) ? JSON.parse(fieldValue) : fieldValue;
        }
        return !_.isEqual(fieldValue, field.value);
      })
      .map(field => ({
        key: field.key,
        value:
          field.type === 'object' || field.type === 'dict' || field.type === 'list'
            ? JSON.parse(formValues[field.key])
            : formValues[field.key],
      }));
    const userChatItem = chatList[chatItemIndex - 1];
    const reqBody: any = {};
    if (userChatItem.fileList) {
      reqBody.selected_files = userChatItem.fileList.map((item: any) => ({
        file_name: item.container_path,
      }));
    }
    sendChat({
      body: {
        ...reqBody,
        interruptAction: 'confirm',
        interruptModifiedArgs: args,
      },
    });
  };

  const renderMessageTip = () => {
    return (
      <div className="dip-mb-12">
        <Alert
          message={interrupt!.data.interrupt_config.confirmation_message || '请确认该工具下列输入参数是否正确'}
          type="info"
          showIcon
        />
      </div>
    );
  };

  const getInitialValue = (item: any) => {
    if (item.type === 'object' || item.type === 'dict' || item.type === 'list') {
      return JSON.stringify(item.value, null, 2);
    }
    return item.value;
  };

  return (
    <div>
      {renderMessageTip()}
      <Form colon form={form} layout="vertical" onFinish={updateInterrupt}>
        {fields.map((item: FieldType) => (
          <FormItem key={item.key} label={item.key} name={item.key} initialValue={getInitialValue(item)}>
            {renderInput(item)}
          </FormItem>
        ))}
        <div className="dip-flex-space-between">
          <span />
          <span>
            <Button
              type="default"
              onClick={() => {
                sendChat({
                  body: {
                    interruptAction: 'skip',
                  },
                });
              }}
            >
              跳过该{SKILL_TYPE[skillInfo.type]}
            </Button>
            <Button
              className={'dip-ml-8'}
              type="primary"
              onClick={() => {
                form.submit();
              }}
            >
              继续执行
            </Button>
          </span>
        </div>
      </Form>
    </div>
  );
};

export default InterruptFormPanel;
