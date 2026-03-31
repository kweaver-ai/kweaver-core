import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Modal, Button, message, Spin, Empty, Checkbox, Tooltip } from 'antd';
import EntityDetail from '../EntityDetail';
import { getKnowledgeNetworkEntityList } from '../utils';
import type { Entity } from '../types';
import './style.less';

interface KnowledgeNetworkResultRangeProps {
  networkId: string;
  resultRangeOptions: string[]; // 主界面已选中的实体键（可选范围）
  resultRange: string[]; // 当前的结果范围（resultRangeOptions的子集）
  onCancel: () => void;
  onSave: (resultRange: string[]) => void;
}

const KnowledgeNetworkResultRange: React.FC<KnowledgeNetworkResultRangeProps> = ({
  networkId,
  resultRangeOptions = [], // 从主界面传入的已选实体
  resultRange = [], // 当前的结果范围
  onCancel,
  onSave,
}) => {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedResultRange, setSelectedResultRange] = useState<string[]>(resultRange);
  const [loading, setLoading] = useState(false);
  const [selectedEntityTitle, setSelectedEntityTitle] = useState<string>('');

  // 获取实体列表
  useEffect(() => {
    if (networkId) {
      loadEntities(networkId);
    }
  }, [networkId]);

  const loadEntities = async (networkId: string) => {
    setLoading(true);
    try {
      const entityList = await getKnowledgeNetworkEntityList(networkId);
      setEntities(entityList.filter(entity => resultRangeOptions.includes(entity.title)));
    } catch {
      message.error(intl.get('dataAgent.fetchEntityListFail'));
    } finally {
      setLoading(false);
    }
  };

  // 处理结果范围选择
  const handleResultRangeCheck = (name: string, checked: boolean) => {
    const newResultRange = checked ? [...selectedResultRange, name] : selectedResultRange.filter(key => key !== name);
    setSelectedResultRange(newResultRange);
  };

  // 保存时返回选中的结果范围
  const handleSave = () => {
    onSave(selectedResultRange);
    message.success(intl.get('dataAgent.resultRangeSetSuccess'));
    onCancel();
  };

  return (
    <Modal
      title={intl.get('dataAgent.specifyResultRange')}
      open={true}
      onCancel={onCancel}
      width={1200}
      centered
      footer={[
        <Button key="save" type="primary" className="dip-min-width-72" onClick={handleSave}>
          {intl.get('dataAgent.config.save')}
        </Button>,
        <Button key="cancel" className="dip-min-width-72" onClick={onCancel}>
          {intl.get('dataAgent.cancel')}
        </Button>,
      ]}
    >
      <div className="result-range-content">
        {/* 左侧实体列表 */}
        <div className="entity-list-panel">
          <Spin spinning={loading}>
            <div className="entity-list-container">
              {entities.length > 0 ? (
                <>
                  <Checkbox
                    className="dip-pl-16 dip-pt-8 dip-pb-8"
                    indeterminate={
                      selectedResultRange.length > 0 && selectedResultRange.length !== resultRangeOptions.length
                    }
                    checked={selectedResultRange.length > 0 && selectedResultRange.length === resultRangeOptions.length}
                    onChange={e => {
                      if (e.target.checked) {
                        setSelectedResultRange(resultRangeOptions);
                      } else {
                        setSelectedResultRange([]);
                      }
                    }}
                  >
                    {intl.get('dataAgent.selectAll')}
                  </Checkbox>
                  <div className="entity-list">
                    {entities.map(entity => {
                      const title = `${entity.alias}（${entity.title}）`;

                      return (
                        <div key={entity.key} className="entity-item dip-flex dip-gap-4">
                          <Checkbox
                            checked={selectedResultRange.includes(entity.title)}
                            onChange={e => handleResultRangeCheck(entity.title, e.target.checked)}
                          />
                          <Tooltip title={title} placement="right">
                            <div
                              className={classNames(
                                'entity-content dip-ellipsis dip-pl-8 dip-pr-8 dip-border-radius-6 dip-pointer dip-h-24',
                                {
                                  'selected-entity-content': selectedEntityTitle === entity.title,
                                }
                              )}
                              onClick={() => setSelectedEntityTitle(entity.title)}
                            >
                              <span
                                className="entity-color-dot dip-border-radius-full dip-mr-8 dip-display-inline-block"
                                style={{
                                  width: '8px',
                                  height: '8px',
                                  backgroundColor: entity.color || '#1890ff',
                                }}
                              />

                              <span className="entity-name">{title}</span>
                            </div>
                          </Tooltip>
                        </div>
                      );
                    })}
                  </div>
                </>
              ) : (
                !loading && (
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={intl.get('dataAgent.noSelectableEntities')}
                  />
                )
              )}
            </div>
          </Spin>
        </div>

        <EntityDetail networkId={networkId} entityTitle={selectedEntityTitle} />
      </div>
    </Modal>
  );
};

export default KnowledgeNetworkResultRange;
