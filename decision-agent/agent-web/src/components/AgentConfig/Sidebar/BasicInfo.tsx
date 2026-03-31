import React, { useState } from 'react';
import intl from 'react-intl-universal';
import { Input, Modal, message, Flex } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { validateName } from '@/utils/validators';
import BasicInfoIcon from '@/assets/icons/base-info.svg';
import { useAgentConfig } from '../AgentConfigContext';
import AgentIcon, { AVATAR_OPTIONS } from '@/components/AgentIcon';
import SectionPanel from '../common/SectionPanel';
import styles from './BasicInfo.module.less';

const { TextArea } = Input;
const MAX_NAME_LENGTH = 50;
const MAX_DESCRIPTION_LENGTH = 200;

const BasicInfo: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [isAvatarModalVisible, setIsAvatarModalVisible] = useState(false);
  const [selectedAvatar, setSelectedAvatar] = useState<{ type: number; value: string } | null>(null);
  const [nameError, setNameError] = useState<string>('');
  const [profileError, setProfileError] = useState<string>('');
  const [isHovering, setIsHovering] = useState(false);
  const [expanded, setExpanded] = useState(true);

  // 检查字段是否可编辑
  const canEditName = actions.canEditField('name');
  const canEditProfile = actions.canEditField('profile');
  const canEditAvatar = actions.canEditField('avatar');

  const showAvatarModal = () => {
    if (!canEditAvatar) return;
    // 默认选中当前头像
    let currentAvatar = AVATAR_OPTIONS[0];
    if (state.avatar_type && state.avatar) {
      const current = AVATAR_OPTIONS.find(option => option.type === state.avatar_type && option.value === state.avatar);
      if (current) {
        currentAvatar = current;
      }
    }
    setSelectedAvatar(currentAvatar);
    setIsAvatarModalVisible(true);
  };

  const handleAvatarSelect = (avatar: { type: number; value: string }) => {
    setSelectedAvatar(avatar);
  };

  const handleAvatarModalOk = () => {
    if (selectedAvatar) {
      // 更新头像到context，同时传递avatarType
      actions.updateBasicInfo(state.name, state.profile, selectedAvatar.value, selectedAvatar.type);
      message.success(intl.get('dataAgent.config.avatarUpdated'));
    }
    setIsAvatarModalVisible(false);
  };

  const handleAvatarModalCancel = () => {
    setIsAvatarModalVisible(false);
  };

  const handleToggle = () => {
    setExpanded(!expanded);
  };

  // 验证名称是否包含非法字符
  const checkName = (name: string): boolean => {
    if (!name || name.trim() === '') {
      setNameError(intl.get('dataAgent.config.nameEmptyError'));
      return false;
    }

    if (!validateName(name)) {
      setNameError(intl.get('dataAgent.config.namingConventionRule'));
      return false;
    }

    setNameError('');

    return true;
  };

  // 验证简介是否不为空
  const validateProfile = (profile: string): boolean => {
    if (!profile || profile.trim() === '') {
      setProfileError(intl.get('dataAgent.config.profileEmptyError'));
      return false;
    }

    setProfileError('');
    return true;
  };

  // 处理名称变更
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!canEditName) return;
    const newName = e.target.value.slice(0, MAX_NAME_LENGTH);

    // 验证输入的名称
    checkName(newName);

    // 仍然更新状态，但在保存时会进行完整验证
    actions.updateBasicInfo(newName, state.profile, state.avatar, state.avatar_type);
  };

  // 处理描述变更
  const handleDescriptionChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    if (!canEditProfile) return;
    const newDescription = e.target.value.slice(0, MAX_DESCRIPTION_LENGTH);

    // 验证简介
    validateProfile(newDescription);

    // 仍然更新状态，但在保存时会进行完整验证
    actions.updateBasicInfo(state.name, newDescription, state.avatar, state.avatar_type);
  };

  return (
    <SectionPanel
      title={intl.get('dataAgent.config.baseInfo')}
      icon={<BasicInfoIcon />}
      isExpanded={expanded}
      onToggle={handleToggle}
      className="dip-border-line-b"
    >
      <div className={styles['form-row']}>
        <div className={styles['form-column']}>
          <div style={{ marginBottom: 16 }}>
            <div className="dip-mb-12">
              <label className="dip-required">{intl.get('dataAgent.config.name')}</label>
            </div>
            <Input
              value={state.name}
              onChange={handleNameChange}
              placeholder={intl.get('dataAgent.config.namePlaceholder')}
              maxLength={MAX_NAME_LENGTH}
              disabled={!canEditName}
              status={nameError ? 'error' : ''}
              suffix={
                <span className={styles['char-count']}>
                  {state.name?.length || 0}/{MAX_NAME_LENGTH}
                </span>
              }
            />
            {nameError && <div className={styles['error-message']}>{nameError}</div>}
          </div>

          <div>
            <div className="dip-mb-12">
              <label className="dip-required">{intl.get('dataAgent.config.profile')}</label>
            </div>
            <TextArea
              value={state.profile}
              onChange={handleDescriptionChange}
              placeholder={intl.get('dataAgent.config.profilePlaceholder')}
              maxLength={MAX_DESCRIPTION_LENGTH}
              disabled={!canEditProfile}
              autoSize={{ minRows: 4, maxRows: 7 }}
              showCount={{
                formatter: ({ count, maxLength }) => (
                  <span className={styles['inner-count']}>
                    {count}/{maxLength}
                  </span>
                ),
              }}
              status={profileError ? 'error' : ''}
            />
            {profileError && <div className={styles['error-message']}>{profileError}</div>}
          </div>
        </div>
        <div
          className={styles['avatar-wrapper']}
          onClick={showAvatarModal}
          onMouseEnter={() => canEditAvatar && setIsHovering(true)}
          onMouseLeave={() => setIsHovering(false)}
          style={{ cursor: canEditAvatar ? 'pointer' : 'default' }}
        >
          <AgentIcon avatar_type={state.avatar_type} avatar={state.avatar} size={114} name={state.name} />
          {canEditAvatar && (
            <div className={styles['avatar-overlay']} style={{ opacity: isHovering ? 1 : 0 }}>
              <div>
                <div style={{ textAlign: 'center', marginBottom: '5px' }}>
                  <EditOutlined style={{ fontSize: '18px' }} />
                </div>
                {intl.get('dataAgent.config.changeAvatar')}
              </div>
            </div>
          )}
        </div>
      </div>

      <Modal
        title={intl.get('dataAgent.config.selectAvatar')}
        centered
        open={isAvatarModalVisible}
        onOk={handleAvatarModalOk}
        onCancel={handleAvatarModalCancel}
        footer={(_, { OkBtn, CancelBtn }) => (
          <>
            <OkBtn />
            <CancelBtn />
          </>
        )}
      >
        <Flex wrap="wrap" gap="small">
          {AVATAR_OPTIONS.map(avatar => (
            <div
              key={avatar.value}
              style={{
                cursor: 'pointer',
                padding: '10px',
                borderRadius: '50%',
                border: selectedAvatar?.value === avatar.value ? '2px solid #1890ff' : '2px solid transparent',
              }}
              onClick={() => handleAvatarSelect(avatar)}
            >
              <AgentIcon avatar_type={avatar.type} avatar={avatar.value} name={state.name} size={64} />
            </div>
          ))}
        </Flex>
      </Modal>
    </SectionPanel>
  );
};

export default BasicInfo;
