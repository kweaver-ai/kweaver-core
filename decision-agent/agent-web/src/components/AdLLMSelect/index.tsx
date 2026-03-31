import React, { useEffect, useMemo, useRef, useState } from 'react';
import _ from 'lodash';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { Col, Empty, InputNumber, message, Row, Select, SelectProps, Switch, Tooltip } from 'antd';
import { QuestionCircleFilled } from '@ant-design/icons';
import { getModelList as getModelListReq, testModelConnection } from '@/apis/model-manager';
import IconFont from '@/components/IconFont';
import kongImg from '@/assets/icons/kong.svg';
import { MODEL_ICON } from './model-icon';
import ExplainTip from '@/components/ExplainTip';
import { useLatestState, useDeepCompareEffect } from '@/hooks';
import { getLLM_CONFIG_PARAM } from './enum';
import './style.less';

export type LLmModelType = {
  icon?: string; // 模型的icon
  id: string; // 模型的id
  name: string; // 模型的名字
  // 以下是模型的参数配置, auto 是智能调参开启的参数
  temperature: any;
  top_p: any;
  top_k: any;
  frequency_penalty: any;
  presence_penalty: any;
  max_tokens: any;
  context_length?: number;
  old_context_length?: number;
};

interface AdLLmModelSelectProps extends Omit<SelectProps, 'value' | 'onChange'> {
  readonly?: boolean;
  value?: LLmModelType;
  onChange?: (value?: LLmModelType) => void;
  defaultSelectedFirstOption?: boolean; // 是否默认选择第一个LLM
  defaultSelectedModel?: string; // 是否默认选择指定类型的一个模型 优先级比defaultSelectedFirstOption高，配置了defaultSelectedModel之后，如果没找到，默认选择能选的第一个 （该字段对应的值是大模型介入菜单列表的  模型 列）
  queryOnFocus?: boolean; // 是否在聚焦的时候 查询一次最新的大模型数据
  testModel?: boolean; // 是否显示测试按钮
  contextVisible?: boolean; // 是否显示context
  dropdownPanelType?: 'common' | 'smart-param'; // 下拉面板的样式，common是普通样式，smart-param是智能调参样式
  bordered?: boolean; // 是否显示边框
  agentTemplateView?: boolean;
}

const getDefaultModelPara = (max_model_len: number) => ({
  temperature: [0, 2, 1],
  top_p: [0, 1, 1],
  presence_penalty: [-2, 2, 0],
  frequency_penalty: [-2, 2, 0],
  max_tokens: [10, max_model_len ? max_model_len * 1000 : 32000, 1000],
  top_k: [1, 1000, 1],
});

const AdLLMSelect: React.FC<AdLLmModelSelectProps> = props => {
  const {
    readonly = false,
    className,
    value,
    onChange,
    placeholder = intl.get('agentCommonConfig.pleaseSelectModel'),
    defaultSelectedFirstOption = false,
    agentTemplateView = false,
    defaultSelectedModel,
    queryOnFocus = false,
    testModel = false,
    disabled = false,
    bordered = true,
    contextVisible = false,
    dropdownPanelType = 'common',
    ...restSelectProps
  } = props;

  const [llmList, setLLMList] = useLatestState<any[]>([]);
  const [open, setOpen] = useState<boolean>(false);
  const dropdownRef = useRef<any>();
  useEffect(() => {
    getModelList();
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current?.contains(e.target)) {
        return;
      }
      setOpen(false);
    };
    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  }, []);

  useDeepCompareEffect(() => {
    if (value && value.name && !value.icon) {
      const model = llmList.find(item => item.model_name === value.name)!;
      if (model) {
        onChange?.({
          ...value,
          icon: model.icon,
        });
      }
      return;
    }

    if (value && contextVisible) {
      const modelFind = _.find(_.cloneDeep(llmList), (modelItem: any) => modelItem?.model_name === value?.name);

      let changeModelValue: any = value;
      // value.context_length为0-升级上来  为''自己清空的
      if (value.context_length === 0) {
        changeModelValue = {
          ...changeModelValue,
          context_length: Math.ceil(modelFind?.max_model_len / 4),
          old_context_length: modelFind?.max_model_len,
        };
      } else if (value.context_length) {
        changeModelValue = {
          ...changeModelValue,
          old_context_length: modelFind?.max_model_len,
        };
      } else {
        changeModelValue = {
          ...changeModelValue,
          context_length: value?.context_length,
          old_context_length: modelFind?.max_model_len,
        };
      }
      onChange?.({ ...changeModelValue });
    }

    if (!value || !value.name) {
      onHandleDefaultModelConfig();
    }
  }, [value, llmList, defaultSelectedModel]);

  /**
   * 赋默认值
   */
  const onHandleDefaultModelConfig = async () => {
    if (defaultSelectedModel || defaultSelectedFirstOption) {
      let defaultModel: any;
      if (defaultSelectedModel) {
        defaultModel = llmList.find(item => item.model === defaultSelectedModel);
      }
      if (!defaultModel || defaultSelectedFirstOption) {
        defaultModel = llmList[0];
      }
      if (defaultModel) {
        const model_para = getDefaultModelPara(defaultModel.max_model_len);

        const modelFindLen = _.find(_.cloneDeep(llmList), item => item.model_name === defaultModel?.model_name);
        let commonConfig: any = {
          icon: defaultModel?.icon,
          id: defaultModel?.model_id,
          name: defaultModel?.model_name,
          temperature: value?.temperature ?? model_para?.temperature[2],
          top_p: value?.top_p ?? model_para?.top_p[2],
          max_tokens: value?.max_tokens ?? model_para?.max_tokens[2],
          top_k: value?.top_k ?? model_para?.top_k[2],
          presence_penalty: value?.presence_penalty ?? model_para?.presence_penalty[2],
          frequency_penalty: value?.frequency_penalty ?? model_para?.frequency_penalty[2],
        };
        if (dropdownPanelType === 'smart-param') {
          commonConfig = {
            ...commonConfig,
            temperature: 'auto',
            top_p: 'auto',
            max_tokens: 'auto',
            top_k: 'auto',
            presence_penalty: 'auto',
            frequency_penalty: 'auto',
          };
        }
        if (contextVisible) {
          commonConfig = {
            ...commonConfig,
            context_length: Math.ceil(modelFindLen?.max_model_len / 4),
            old_context_length: modelFindLen?.max_model_len,
          };
        }
        onChange?.(commonConfig);
      }
    }
  };

  /**
   * 获取模型
   */
  const getModelList = async () => {
    const { data } = await getModelListReq({ page: 1, size: 100 });
    setLLMList(data || []);
  };

  const prefixCls = 'AdLLMSelect';

  const getModelIcon = (icon: string) => {
    return MODEL_ICON[icon];
  };

  const modelOptions = useMemo(() => {
    return llmList.map((item: any) => {
      const Icon = getModelIcon(item.icon);
      return {
        label: (
          <div className="dip-flex-align-center">
            {Icon ? <Icon /> : null}
            <span
              className={classNames('dip-ellipsis dip-flex-item-full-width', {
                'dip-ml-8': Icon,
              })}
              title={item.model_name}
            >
              {item.model_name}
            </span>
          </div>
        ),
        value: item.model_name,
      };
    });
  }, [llmList]);

  const selectedModelParamConfig: any = useMemo(() => {
    const target = llmList.find(item => item.model_name === value?.name);
    if (target) {
      const model_para = getDefaultModelPara(target.max_model_len);

      return {
        temperature: {
          min: model_para.temperature[0],
          max: model_para.temperature[1],
          value: value?.temperature ?? model_para.temperature[2],
          defaultValue: model_para.temperature[2],
        },
        top_p: {
          min: model_para.top_p[0],
          max: model_para.top_p[1],
          value: value?.top_p ?? model_para.top_p[2],
          defaultValue: model_para.top_p[2],
        },
        max_tokens: {
          min: model_para.max_tokens[0],
          max: model_para.max_tokens[1],
          value: value?.max_tokens ?? model_para.max_tokens[2],
          defaultValue: model_para.max_tokens[2],
        },
        top_k: {
          min: model_para.top_k[0],
          max: model_para.top_k[1],
          value: value?.top_k ?? model_para.top_k[2],
          defaultValue: model_para.top_k[2],
        },
        presence_penalty: {
          min: model_para.presence_penalty[0],
          max: model_para.presence_penalty[1],
          value: value?.presence_penalty ?? model_para.presence_penalty[2],
          defaultValue: model_para.presence_penalty[2],
        },
        frequency_penalty: {
          min: model_para.frequency_penalty[0],
          max: model_para.frequency_penalty[1],
          value: value?.frequency_penalty ?? model_para.frequency_penalty[2],
          defaultValue: model_para.frequency_penalty[2],
        },
      };
    }
    return null;
  }, [value, llmList]);

  const paramChange = (paramValue: number | 'auto', paramKey: string) => {
    onChange?.({
      ...value,
      [paramKey]: paramValue,
    } as any);
  };

  const modelSelect = (name: string) => {
    const model = llmList.find(item => item.model_name === name)!;
    const findModel = llmList.find((modelItem: any) => modelItem.model_name === name)!;
    const model_para = getDefaultModelPara(model?.max_model_len);
    let defaultModel: any = {
      icon: model.icon,
      id: model.model_id,
      name: model.model_name,
      temperature: value?.temperature ?? model_para.temperature[2],
      top_p: value?.top_p ?? model_para.top_p[2],
      max_tokens: value?.max_tokens ?? model_para.max_tokens[2],
      top_k: value?.top_k ?? model_para.top_k[2],
      presence_penalty: value?.presence_penalty ?? model_para.presence_penalty[2],
      frequency_penalty: value?.frequency_penalty ?? model_para.frequency_penalty[2],
    };
    if (dropdownPanelType === 'smart-param') {
      defaultModel = {
        ...defaultModel,
        temperature: 'auto',
        top_p: 'auto',
        max_tokens: 'auto',
        top_k: 'auto',
        presence_penalty: 'auto',
        frequency_penalty: 'auto',
      };
    }
    if (contextVisible) {
      defaultModel = {
        ...defaultModel,
        context_length: Math.ceil(findModel?.max_model_len / 4),
        old_context_length: findModel?.max_model_len,
      };
    }
    onChange?.(defaultModel);
  };

  const renderDropdownPanel = () => {
    if (dropdownPanelType === 'common') {
      return (
        <>
          <div className="dip-mb-20 dip-flex-align-center" style={{ marginTop: 10 }}>
            {intl.get('agentCommonConfig.configurationParameters')}
            <div className="dip-flex-item-full-width dip-ml-8" style={{ height: 1, background: '#e5e5e5' }} />
          </div>
          {getLLM_CONFIG_PARAM().map((item: any) => {
            const min = selectedModelParamConfig ? selectedModelParamConfig[item.key].min : 0;
            const max = selectedModelParamConfig ? selectedModelParamConfig[item.key].max : 0;
            const value = selectedModelParamConfig ? selectedModelParamConfig[item.key].value : 0;
            return (
              <div key={item.key} className="dip-mb-16">
                <div className="dip-flex-space-between">
                  <div className="dip-flex-align-center">
                    <span>{item.label}</span>
                    <ExplainTip title={item.tip} />
                  </div>
                  <div className="dip-flex-align-center">
                    <InputNumber
                      disabled={disabled}
                      size="small"
                      step={item.step}
                      precision={item.precision}
                      min={min}
                      max={max}
                      value={value}
                      onChange={data => {
                        paramChange(data, item.key);
                      }}
                    />
                  </div>
                </div>
              </div>
            );
          })}
        </>
      );
    }

    if (dropdownPanelType === 'smart-param') {
      return (
        <>
          <Row justify="space-between" className="dip-mb-8 dip-mt-16">
            <Col className="dip-c-subtext" span={14}>
              {intl.get('federalAi.config.parameter')}
            </Col>
            <Col className="dip-c-subtext dip-flex-center" span={4}>
              {intl.get('federalAi.config.smartParameter')}
            </Col>
            <Col className="dip-c-subtext" span={6}>
              {intl.get('federalAi.config.fixParam')}
            </Col>
          </Row>
          {getLLM_CONFIG_PARAM().map((item: any) => {
            const min = selectedModelParamConfig ? selectedModelParamConfig[item.key].min : 0;
            const max = selectedModelParamConfig ? selectedModelParamConfig[item.key].max : 0;
            const value = selectedModelParamConfig ? selectedModelParamConfig[item.key].value : 0;
            const defaultValue = selectedModelParamConfig ? selectedModelParamConfig[item.key].defaultValue : 0;
            return (
              <Row justify="space-between" key={item.key} className="dip-mb-16">
                <Col span={14}>
                  <div className="dip-flex-align-center">
                    <span>{item.label}</span>
                    <ExplainTip title={item.tip} />
                  </div>
                </Col>
                <Col span={4} className="dip-flex-center">
                  <Switch
                    disabled={disabled}
                    checked={value === 'auto'}
                    size="small"
                    onChange={checked => {
                      if (checked) {
                        paramChange('auto', item.key);
                      } else {
                        paramChange(defaultValue, item.key);
                      }
                    }}
                  />
                </Col>
                <Col span={6}>
                  {value === 'auto' ? (
                    <InputNumber size="small" disabled />
                  ) : (
                    <InputNumber
                      disabled={disabled}
                      size="small"
                      step={item.step}
                      precision={item.precision}
                      min={min}
                      max={max}
                      value={value}
                      onChange={data => {
                        paramChange(data, item.key);
                      }}
                    />
                  )}
                </Col>
              </Row>
            );
          })}
        </>
      );
    }
  };

  const dropdownRender: SelectProps['dropdownRender'] = menu => {
    return (
      <div
        ref={dropdownRef}
        className={classNames(`${prefixCls}-content`, {
          [`${prefixCls}-content-smartParam`]: dropdownPanelType === 'smart-param',
        })}
        onClick={e => {
          e.stopPropagation();
        }}
      >
        <Select
          value={value?.name}
          notFoundContent={<Empty image={kongImg} description={intl.get('global.noData')} />}
          className="dip-w-100"
          placeholder={intl.get('global.pleaseSelect')}
          options={modelOptions}
          onChange={modelSelect}
          disabled={disabled}
        />

        {selectedModelParamConfig && renderDropdownPanel()}
      </div>
    );
  };

  const handleTestModel = async (modelValue?: LLmModelType) => {
    if (!modelValue) {
      return;
    }
    const { data: modelList } = (await getModelListReq({ page: 1, size: 100 })) || {};
    const model = _.filter(modelList, item => item?.model_name === modelValue?.name)?.[0];
    if (!model) return;

    const testData = {
      // model_config: {
      //   ...model?.model_config,
      //   ..._.omit(model, 'model_config', 'icon', 'model_id', 'model_series', 'model_url'),
      // },
      // model_series: model?.model_series,
      model_id: model?.model_id,
    };

    try {
      const { status } = (await testModelConnection(testData)) || {};
      if (status) {
        message.success(intl.get('global.testSuccessful'));
      }
    } catch (err) {
      const { description } = err?.response || err?.data || err || {};
      description && message.error(description);
    }
  };

  /**
   * 输入框数据变化
   */
  const onInputNumberChange = (valueData: any) => {
    const contextLengthValue = !valueData
      ? ''
      : _.includes(valueData, '.')
        ? parseFloat(valueData)
        : parseInt(valueData);
    onChange?.({
      ...value,
      context_length: contextLengthValue,
    } as any);
  };

  return (
    <div className={classNames(prefixCls, className, 'dip-flex-align-center dip-w-100')}>
      <Select
        bordered={bordered}
        disabled={false}
        dropdownMatchSelectWidth={false}
        className={'dip-w-100'}
        open={open}
        value={value?.name}
        placeholder={placeholder}
        options={modelOptions}
        dropdownRender={value?.name ? dropdownRender : undefined}
        suffixIcon={
          <>
            {/* {testModel && */}
            {(agentTemplateView
              ? testModel && _.find(_.cloneDeep(llmList), (modelItem: any) => modelItem?.model_name === value?.name)
              : testModel) &&
              !readonly && (
                <Tooltip title={intl.get('global.test')}>
                  <IconFont
                    className="dip-pointer"
                    type="icon-a-ceshimodel"
                    onClick={e => {
                      e.stopPropagation();
                      handleTestModel(value);
                    }}
                  />
                </Tooltip>
              )}
            <Tooltip title={intl.get('global.settings')}>
              <IconFont className="dip-c-text-lower" type="icon-setting" />
            </Tooltip>
          </>
        }
        onChange={modelSelect}
        onClear={() => {
          onChange?.();
        }}
        onClick={e => {
          if (queryOnFocus) {
            getModelList();
          }
          setOpen(true);
          e.stopPropagation();
        }}
        {...restSelectProps}
      />
      {contextVisible ? (
        <div className={classNames(`${prefixCls}-model-test-context`, 'dip-flex-align-center dip-ml-24')}>
          <div className="dip-ellipsis dip-flex-align-center">
            Context Window-Retriever
            <Tooltip title={intl.get('agentCommonConfig.contextDesc')} placement="top">
              <QuestionCircleFilled className="dip-c-watermark dip-ml-4 dip-mr-8" />
            </Tooltip>
            ：
          </div>
          <InputNumber
            className="dip-ml-12"
            min={1}
            style={{ width: 65 }}
            max={value?.old_context_length}
            precision={0}
            controls={false}
            value={value?.context_length}
            onChange={(valueData: any) => onInputNumberChange(valueData)}
            disabled={disabled}
            placeholder={
              value?.old_context_length && !value?.context_length && value?.context_length !== 0
                ? `1~${value?.old_context_length}`
                : ''
            }
            onBlur={(e: any) => {
              // 值不存在，则显示默认值
              if (!e?.target?.value) {
                onInputNumberChange(value?.old_context_length);
              }
            }}
          />
          <div className="dip-ml-12">K</div>
        </div>
      ) : null}
    </div>
  );
};

export default AdLLMSelect;
