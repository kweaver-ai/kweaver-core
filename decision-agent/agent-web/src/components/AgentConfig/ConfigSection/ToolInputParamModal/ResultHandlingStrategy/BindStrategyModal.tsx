import { memo, useState, useEffect } from 'react';
import classNames from 'classnames';
import { Modal, Select, message } from 'antd';
import {
  getToolProcessCategories,
  getToolProcessStrategiesByCategoryId,
  type ToolProcessCategoryType,
  type ToolProcessStrategyType,
  type ResultProcessStrategyType,
} from '@/apis/agent-factory';
import styles from './BindStrategyModal.module.less';
import intl from 'react-intl-universal';

interface Props {
  onCancel: () => void;
  onConfirm: (params: ResultProcessStrategyType) => void;
}

const BindStrategyModal = memo(({ onCancel, onConfirm }: Props) => {
  // 分类选项
  const [categories, setCategories] = useState<Array<ToolProcessCategoryType>>([]);
  // 选中的分类
  const [selectedCategory, setSelectedCategory] = useState<ToolProcessCategoryType | undefined>(undefined);
  // 策略选项
  const [strategies, setStrategies] = useState<Array<ToolProcessStrategyType>>([]);
  // 选中的策略
  const [selectedStrategy, setSelectedStrategy] = useState<ToolProcessStrategyType | undefined>(undefined);

  // 获取分类选项
  const fetchCategories = async () => {
    try {
      const { entries } = await getToolProcessCategories();

      setCategories(entries);
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    }
  };

  // 获取策略选项
  const fetchStrategies = async (categoryId: string) => {
    try {
      const { entries } = await getToolProcessStrategiesByCategoryId(categoryId);
      setStrategies(entries);
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    }
  };

  useEffect(() => {
    // 获取分类选项
    fetchCategories();
  }, []);

  useEffect(() => {
    if (!selectedCategory) return;

    // 当选中的分类发生变化，获取其对应的策略选项
    fetchStrategies(selectedCategory.id);
    // 清空选中的策略
    setSelectedStrategy(undefined);
  }, [selectedCategory]);

  return (
    <Modal
      centered
      open
      maskClosable={false}
      title={intl.get('dataAgent.bindPolicy')}
      width={466}
      okButtonProps={{ disabled: !(selectedCategory && selectedStrategy), className: 'dip-min-width-72' }}
      cancelButtonProps={{ className: 'dip-min-width-72' }}
      onCancel={onCancel}
      onOk={() => onConfirm({ category: selectedCategory!, strategy: selectedStrategy! })}
      footer={(_, { OkBtn, CancelBtn }: any) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
    >
      <div className={classNames('dip-mt-12 dip-mb-20', styles['content'])}>
        <div className="dip-mb-4 dip-c-black">{intl.get('dataAgent.policyCategory')}</div>
        <div>
          <Select
            className="dip-w-100"
            placeholder={intl.get('dataAgent.selectPolicyCategory')}
            fieldNames={{
              value: 'id',
              label: 'name',
            }}
            options={categories}
            onSelect={(_, option) => setSelectedCategory(option)}
          />
        </div>

        <div className="dip-mt-24 dip-mb-4 dip-c-black">{intl.get('dataAgent.policyName')}</div>
        <Select
          className="dip-w-100"
          placeholder={intl.get('dataAgent.selectPolicyName')}
          fieldNames={{
            value: 'id',
            label: 'name',
          }}
          value={selectedStrategy?.id}
          options={strategies}
          onSelect={(_, option) => setSelectedStrategy(option)}
        />

        {selectedStrategy?.id ? (
          <>
            <div className="dip-mt-24 dip-mb-4 dip-c-black">{intl.get('dataAgent.policyDescription')}</div>
            <div className="dip-pl-12 dip-text-color-75">
              {selectedStrategy?.description || intl.get('dataAgent.noDescriptionAvailable')}
            </div>
          </>
        ) : null}
      </div>
    </Modal>
  );
});

export default BindStrategyModal;
