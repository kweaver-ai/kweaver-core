import { FC, useEffect, useMemo, useState, memo, useRef, useCallback } from 'react';
import { Modal, Tree, TreeDataNode, Radio, InputNumber, Tooltip, Button, message, Input } from 'antd';
import { DownOutlined, QuestionCircleOutlined, SearchOutlined } from '@ant-design/icons';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { last } from 'lodash';
import { copyToBoard } from '@/utils/handle-function';
import { getSelfConfigSchema } from '@/apis/agent-factory';
import { SelfConfigSchemaType } from '@/apis/agent-factory/type';
import DipIcon from '@/components/DipIcon';
import { selfConfigConst } from './types';
import PathExplanation from './PathExplanation';
import styles from './SelfConfigVarSelector.module.css';

// 数组选择选项
enum ArraySelectOptionEnum {
  // 整个数组
  All = 'all',

  // 特定索引的元素
  Index = 'index',
}

enum TypeEnum {
  String = 'string',
  Integer = 'integer',
  Number = 'number',
  Boolean = 'boolean',
  Object = 'object',
  Array = 'array',
}

// 匹配字符串末尾的 [数字] 格式（如 "[123]"）
const endingBracketNumberRegex = /\[\d+\]$/;

interface SelfConfigVarSelectorProps {
  defaultValue?: string;
  onCancel: () => void;
  onConfirm: (path: string) => void;
}

// 从树形数据中查找指定key的节点
const findNodeByKey = (data: any[], targetKey: string): any | null => {
  for (const item of data) {
    if (item.key === targetKey) {
      return item;
    }
    if (item.children && item.children.length) {
      const found = findNodeByKey(item.children, targetKey);
      if (found) return found;
    }
  }
  return null;
};

// 统计字符串中点的数量
const countDots = (str: string) => (str.match(/\./g) || []).length;

// 按点分割生成前缀
function generateDotSeparatedPrefixes(str: string | undefined): string[] {
  if (!str) return [];

  // 按点号分割字符串
  const parts = str.split('.');
  const result = [];

  // 逐步构建每个前缀
  for (let i = 1; i <= parts.length; i++) {
    // 取前i个部分并拼接成字符串
    result.push(
      parts
        .slice(0, i)
        .join('.')
        .replace(/\[\*\]$/, '') // 去除尾部的 [*]
    );
  }

  return result;
}

// 解析数组索引
const parseArrayIndex = (value: string): number => {
  const match = value.match(endingBracketNumberRegex);
  if (!match) return 0;

  const index = parseInt(match[0].replace(/\[|\]/g, ''), 10);
  return isNaN(index) ? 0 : index;
};

// 移除末尾的括号内容（移除[<index>]，移除[*]）
const removeTrailingBrackets = (value: string) => {
  return value.replace(endingBracketNumberRegex, '').replace(/\[\*+\]$/, '');
};

const SelfConfigVarSelector: FC<SelfConfigVarSelectorProps> = ({ defaultValue, onCancel, onConfirm }) => {
  const treeRef = useRef<any>(null);
  const searchHoverKeys = useRef<string[]>([]); // 搜索高亮的keys

  const [searchValue, setSearchValue] = useState(defaultValue || '');
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedNode, setSelectedNode] = useState<any | undefined>(undefined);
  const [selfConfigSchema, setSelfConfigSchema] = useState<SelfConfigSchemaType | undefined>(undefined);

  const copy = useCallback(async (text: string) => {
    if (await copyToBoard(text)) {
      message.success(intl.get('dataAgent.copyKeySuccess'));
    }
  }, []);

  const typeLabels = useMemo(
    () => ({
      [TypeEnum.Object]: intl.get('dataAgent.config.object'),
      [TypeEnum.Integer]: intl.get('dataAgent.integer'),
      [TypeEnum.Number]: intl.get('dataAgent.number'),
      [TypeEnum.Boolean]: intl.get('dataAgent.boolean'),
      [TypeEnum.String]: intl.get('dataAgent.string'),
      [TypeEnum.Array]: intl.get('dataAgent.array'),
    }),
    []
  );

  const treeData = useMemo(() => {
    searchHoverKeys.current = [];

    if (!selfConfigSchema) return [];

    // 类型标签组件
    const TypeTag: FC<{ type: string }> = ({ type }) => (
      <span
        className={classNames('dip-font-10 dip-border-radius-3 dip-ml-10 dip-user-select-none', styles['type-tag'])}
      >
        {typeLabels[type as TypeEnum] || type}
      </span>
    );

    const Description: FC<{ description: string }> = ({ description }) => (
      <span className="dip-ml-10 dip-user-select-none dip-text-color-80 dip-font-11">{description}</span>
    );

    // 提取匹配逻辑
    const getMatchInfo = (key: string, item: any) => {
      if (!searchValue) {
        return { isMatch: false, matchStr: '' };
      }

      // 处理带点的搜索值（key匹配）：使用key前置匹配
      if (searchValue.includes('.')) {
        const filteredSearch = searchValue.replace(endingBracketNumberRegex, '');
        const isMatch = key.startsWith(filteredSearch) && countDots(filteredSearch) === countDots(key);
        const matchStr = isMatch ? last(filteredSearch.split('.')) || '' : '';

        return { isMatch, matchStr };
      }

      // 不带点的搜索值（name匹配）：包含匹配
      return {
        isMatch: item.name.includes(searchValue),
        matchStr: searchValue,
      };
    };

    // 标题渲染逻辑
    const renderTitle = (key: string, item: any, isMatch: boolean, matchStr: string) => {
      const strTitle = item.name;

      // 不匹配或无搜索值时的基础标题
      if (!isMatch || !matchStr) {
        return (
          <span key={key}>
            {strTitle}
            <TypeTag type={item.type} />
            <Description description={item.description} />
          </span>
        );
      }

      // 匹配时的高亮标题
      const index = strTitle.indexOf(matchStr);
      if (index === -1) {
        return (
          <span key={key}>
            {strTitle}
            <TypeTag type={item.type} />
            <Description description={item.description} />
          </span>
        );
      }

      // 添加到搜索高亮键列表
      searchHoverKeys.current = [...searchHoverKeys.current, key];

      const beforeStr = strTitle.substring(0, index);
      const afterStr = strTitle.slice(index + matchStr.length);

      return (
        <span key={key}>
          {beforeStr}
          <span className={styles['search-hover-bg']}>{matchStr}</span>
          {afterStr}
          <TypeTag type={item.type} />
          <Description description={item.description} />
        </span>
      );
    };

    // 递归处理节点的主函数
    const loop = (data: any[], prefixKey: string = '', parent?: any): TreeDataNode[] =>
      data.map(item => {
        const key = prefixKey + item.name;
        const { isMatch, matchStr } = getMatchInfo(key, item);
        const title = renderTitle(key, item, isMatch, matchStr);

        // 处理子节点
        if (item.children) {
          const childPrefix = key + (item.type === TypeEnum.Array ? '[*].' : '.');
          return {
            title,
            key,
            details: { ...item, parent },
            children: loop(item.children, childPrefix, { ...item, key }),
          };
        }

        return {
          title,
          key,
          details: { ...item, parent },
        };
      });

    const data = loop([selfConfigSchema]);

    // 当精准匹配时，自动选中此节点
    if (
      searchHoverKeys.current.length === 1 &&
      searchHoverKeys.current[0] === searchValue.replace(endingBracketNumberRegex, '')
    ) {
      const node = findNodeByKey(data, searchHoverKeys.current[0]);

      if (node) {
        const updatedDetails = { ...node.details };

        // 处理数组类型的索引信息
        if (node.details.type === TypeEnum.Array && endingBracketNumberRegex.test(searchValue)) {
          updatedDetails.arraySelectOption = ArraySelectOptionEnum.Index;
          updatedDetails.arrayIndex = parseArrayIndex(searchValue);
        }

        // 处理数组类型的默认选择选项
        if (node.details.type === TypeEnum.Array && !updatedDetails.arraySelectOption) {
          updatedDetails.arraySelectOption = ArraySelectOptionEnum.All;
        }

        setSelectedNode({
          ...node,
          details: updatedDetails,
        });
      }
    }

    return data;
  }, [searchValue, selfConfigSchema]);

  const path = useMemo(() => {
    if (!selectedNode) return '';

    return selectedNode.details.type === TypeEnum.Array &&
      selectedNode.details.arraySelectOption === ArraySelectOptionEnum.Index
      ? `${selectedNode.key}[${selectedNode.details.arrayIndex || 0}]`
      : selectedNode.key;
  }, [selectedNode]);

  // 获取self_config字段结构
  useEffect(() => {
    const fetchData = async () => {
      try {
        const selfConfigSchema = await getSelfConfigSchema();
        setSelfConfigSchema(selfConfigSchema);
      } catch (ex: any) {
        if (ex?.description) {
          message.error(ex.description);
        }
      }
    };

    fetchData();
  }, []);

  // 设置默认selectedNode
  useEffect(() => {
    // 已有选中节点或无树形数据时直接返回
    if (selectedNode || !treeData.length) return;

    let node;
    const cleanedDefaultValue = defaultValue?.replace(endingBracketNumberRegex, '');

    // 根据defaultValue查找节点
    if (defaultValue && cleanedDefaultValue) {
      node = findNodeByKey(treeData, cleanedDefaultValue);

      // 处理数组类型的索引信息
      if (node && node.details.type === TypeEnum.Array && endingBracketNumberRegex.test(defaultValue)) {
        node.details.arraySelectOption = ArraySelectOptionEnum.Index;
        node.details.arrayIndex = parseArrayIndex(defaultValue);
      }
    }

    // 如果未找到节点，使用第一个节点作为默认值
    if (!node) {
      node = treeData[0];
    }

    // 处理数组类型的默认选择选项
    if (node.details.type === TypeEnum.Array && !node.details.arraySelectOption) {
      setSelectedNode({
        ...node,
        details: { ...node.details, arraySelectOption: ArraySelectOptionEnum.All },
      });
    } else {
      setSelectedNode(node);
    }
  }, [treeData]);

  // 设置默认expandedKeys
  useEffect(() => {
    if (expandedKeys.length || !treeData.length) return;

    setExpandedKeys([treeData[0].key]);
  }, [treeData]);

  // 当搜索时，自动展开父节点
  useEffect(() => {
    if (!searchValue) return;

    const filterSearchValue = removeTrailingBrackets(searchValue);
    const hasDot = searchValue.includes('.');
    const searchValueDotCount = hasDot ? countDots(filterSearchValue) : 0;

    // 收集所有需要展开的键
    const keys = new Set<string>();

    // 递归处理节点的函数
    const processNode = (node: any) => {
      // 检查是否匹配搜索条件
      const isMatch = hasDot
        ? node.key.startsWith(filterSearchValue) && countDots(node.key) === searchValueDotCount
        : node.details.name.includes(filterSearchValue);

      if (isMatch) {
        // 收集所有需要展开的前缀键
        generateDotSeparatedPrefixes(node.details.parent?.key).forEach(key => {
          if (key) keys.add(key);
        });
      }

      // 递归处理子节点
      if (node.children) {
        node.children.forEach(processNode);
      }
    };

    treeData.forEach(processNode);

    // 设置展开的键（确保至少展开第一个节点）
    const newKeys = Array.from(keys).filter(Boolean);
    const newExpandKeys = newKeys.length ? newKeys : [treeData[0]?.key];
    setExpandedKeys(newExpandKeys);

    // 将高亮的key展示在可视区域
    if (newExpandKeys.length && searchHoverKeys.current.length && treeRef.current) {
      treeRef.current.scrollTo({ key: searchHoverKeys.current[0] });
    }
  }, [searchValue, selfConfigSchema]);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchValue(e.target.value);
  }, []);

  const handleTreeSelect = useCallback((_, { selectedNodes }: any) => {
    if (selectedNodes.length === 0) return;

    const node = selectedNodes[0];

    if (node.details.type === TypeEnum.Array) {
      setSelectedNode({
        ...node,
        details: { ...node.details, arraySelectOption: ArraySelectOptionEnum.All },
      });
    } else {
      setSelectedNode(node);
    }
  }, []);

  const handleArrayOptionChange = useCallback((value: ArraySelectOptionEnum) => {
    setSelectedNode((prev: any) => ({
      ...prev,
      details: { ...prev.details, arraySelectOption: value, arrayIndex: 0 },
    }));
  }, []);

  const handleArrayIndexChange = useCallback((value: number | null) => {
    setSelectedNode((prev: any) => ({
      ...prev,
      details: { ...prev.details, arrayIndex: value || 0 },
    }));
  }, []);

  return (
    <Modal
      title={intl.get('dataAgent.selectSelfConfigSubfield')}
      open
      centered
      maskClosable={false}
      width={1000}
      okButtonProps={{ className: 'dip-min-width-72', disabled: !selectedNode }}
      cancelButtonProps={{ className: 'dip-min-width-72' }}
      onOk={() => onConfirm(path)}
      onCancel={onCancel}
      footer={(_, { OkBtn, CancelBtn }) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
    >
      <div className={classNames('dip-flex dip-gap-16', styles['container'])}>
        <div className={classNames('dip-border-radius-6 dip-flex-column', styles['tree-wrapper'])}>
          <Input
            allowClear
            className="dip-mb-12"
            placeholder={intl.get('dataAgent.searchFieldPlaceholder')}
            prefix={<SearchOutlined className="dip-opacity-75" />}
            value={searchValue}
            onChange={handleSearchChange}
          />
          <Tree
            ref={treeRef}
            height={428}
            className={classNames('dip-flex-1', styles['tree'])}
            showLine
            switcherIcon={<DownOutlined />}
            treeData={treeData}
            selectedKeys={selectedNode ? [selectedNode.key] : []}
            expandedKeys={expandedKeys}
            onExpand={setExpandedKeys}
            onSelect={handleTreeSelect}
          />
        </div>
        <div className={classNames('dip-border-radius-6 dip-flex-column dip-gap-8', styles['detail'])}>
          <div className="dip-font-weight-700 dip-user-select-none">{intl.get('dataAgent.fieldDetails')}</div>
          {selectedNode ? (
            <div className={classNames('dip-flex-1 dip-bg-white dip-border-6', styles['infos'])}>
              <div className="dip-font-weight-700 dip-pb-10 dip-mb-10 dip-border-b">{selectedNode.details.name}</div>

              <div className="dip-font-12 dip-text-color-65 dip-mb-8 dip-user-select-none">
                {intl.get('dataAgent.fieldType')}
              </div>
              <div
                className={classNames('dip-mb-16 dip-font-12 dip-border-radius-3 dip-w-fit-content', styles['type'])}
              >
                {typeLabels[selectedNode.details.type as TypeEnum]}
              </div>

              <div className="dip-font-12 dip-text-color-65  dip-mb-8 dip-user-select-none">
                {intl.get('dataAgent.fieldDescription')}
              </div>
              <div className={classNames('dip-mb-16 dip-font-13 dip-border-radius-4', styles['description'])}>
                {selectedNode.details.description}
              </div>

              {selectedNode.details.type === TypeEnum.Array && (
                <>
                  <div className="dip-font-12 dip-text-color-65 dip-mb-8 dip-user-select-none">
                    {intl.get('dataAgent.arraySelectionOptions')}
                  </div>
                  <div className="dip-mb-16">
                    <Radio.Group
                      className={classNames('dip-flex-column dip-gap-8', styles['radio-group'])}
                      value={selectedNode.details.arraySelectOption}
                      onChange={e => handleArrayOptionChange(e.target.value)}
                      options={[
                        {
                          value: ArraySelectOptionEnum.All,
                          label: intl.get('dataAgent.entireArray'),
                        },
                        {
                          value: ArraySelectOptionEnum.Index,
                          label: (
                            <div>
                              <div>{intl.get('dataAgent.specificIndexElement')}</div>
                              {selectedNode.details.arraySelectOption === ArraySelectOptionEnum.Index && (
                                <div className="dip-font-12 dip-text-color-65 dip-mt-8">
                                  {intl.get('dataAgent.arrayIndex')}
                                  <InputNumber
                                    value={selectedNode.details.arrayIndex || 0}
                                    min={0}
                                    size="small"
                                    className="dip-ml-10"
                                    onChange={handleArrayIndexChange}
                                  />
                                  <Tooltip
                                    title={intl.get('dataAgent.getElementAtIndex', {
                                      index: selectedNode.details.arrayIndex + 1,
                                    })}
                                  >
                                    <QuestionCircleOutlined className="dip-ml-6 dip-pointer dip-text-color-45 dip-font-14" />
                                  </Tooltip>
                                </div>
                              )}
                            </div>
                          ),
                        },
                      ]}
                    />
                  </div>
                </>
              )}

              <div className="dip-font-12 dip-text-color-65 dip-mb-8 dip-user-select-none">
                {intl.get('dataAgent.selectPath')}
              </div>
              <div className="dip-flex dip-position-r">
                <div className={classNames('dip-font-12 dip-border-radius-4 dip-flex-1', styles['path'])}>{path}</div>
                <Tooltip title={intl.get('dataAgent.copy')}>
                  <Button
                    size="small"
                    type="text"
                    icon={<DipIcon className="dip-text-color-45" type="icon-dip-copy" />}
                    className="dip-ml-4 dip-position-a"
                    style={{ top: 2, right: 2 }}
                    onClick={() => copy(path)}
                  />
                </Tooltip>
              </div>

              {selectedNode.key === selfConfigConst && <PathExplanation isRoot={true} />}

              {selectedNode.details?.parent?.type === TypeEnum.Array && <PathExplanation />}
            </div>
          ) : null}
        </div>
      </div>
    </Modal>
  );
};

export default memo(SelfConfigVarSelector);
