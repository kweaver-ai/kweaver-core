import React, { useEffect } from 'react';
import DipModal from '@/components/DipModal';
import { Form, Input, message } from 'antd';
import { updateConversation } from '@/apis/super-assistant';
import { useDipChatStore } from '@/components/DipChat/store';
import intl from 'react-intl-universal';

const FormItem = Form.Item;
const EditNameModal = ({ onClose, refreshList, name, conversationKey }: any) => {
  const {
    dipChatStore: { agentAppKey },
  } = useDipChatStore();
  const [form] = Form.useForm();
  const nameValue = Form.useWatch('name', form);
  useEffect(() => {
    form.setFieldValue('name', name);
  }, []);
  const onFinish = async ({ name }: any) => {
    const res = await updateConversation(agentAppKey, conversationKey, {
      title: name,
    });
    if (res) {
      message.success(intl.get('dipChat.modifySuccess'));
      onClose();
      refreshList();
    }
  };
  return (
    <>
      <DipModal
        open
        title={intl.get('dipChat.editConversationName')}
        onCancel={onClose}
        onOk={() => {
          form.submit();
        }}
        okButtonProps={{
          disabled: nameValue === name,
        }}
      >
        <Form form={form} onFinish={onFinish}>
          <FormItem
            label=""
            name="name"
            rules={[
              { required: true, message: intl.get('dipChat.conversationNameRequired') },
              { max: 50, message: intl.get('global.lenErr', { len: 50 }) },
            ]}
          >
            <Input placeholder={intl.get('dipChat.enterConversationName')} />
          </FormItem>
        </Form>
      </DipModal>
    </>
  );
};

export default ({ open, ...restProps }: any) => {
  return open && <EditNameModal {...restProps} />;
};
