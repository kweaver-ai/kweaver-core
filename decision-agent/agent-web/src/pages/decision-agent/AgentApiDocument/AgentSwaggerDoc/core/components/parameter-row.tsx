import React, { Component } from 'react';
import { List, Map } from 'immutable';
import PropTypes from 'prop-types';
import { Tooltip } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';
import ImPropTypes from 'react-immutable-proptypes';
import win from '../window';
import { getCommonExtensions, getExtensions, getSampleSchema, isEmptyValue, numberToString, stringify } from '../utils';
import getParameterSchema from '../../helpers/get-parameter-schema';
import ModelExample from './model-example';
import ParamBody from './param-body';
import { JsonSchemaForm } from './json-schema-components';
import { FilterEnum } from '../../types';

export default class ParameterRow extends Component {
  static propTypes = {
    onChange: PropTypes.func.isRequired,
    param: PropTypes.object.isRequired,
    rawParam: PropTypes.object.isRequired,
    getComponent: PropTypes.func.isRequired,
    fn: PropTypes.object.isRequired,
    isExecute: PropTypes.bool,
    onChangeConsumes: PropTypes.func.isRequired,
    specSelectors: PropTypes.object.isRequired,
    specActions: PropTypes.object.isRequired,
    pathMethod: PropTypes.array.isRequired,
    getConfigs: PropTypes.func.isRequired,
    specPath: ImPropTypes.list.isRequired,
    oas3Actions: PropTypes.object.isRequired,
    oas3Selectors: PropTypes.object.isRequired,
    filter: PropTypes.string,
  };

  constructor(props, context) {
    super(props, context);

    this.setDefaultValue();
  }

  UNSAFE_componentWillReceiveProps(props) {
    const { specSelectors, pathMethod, rawParam } = props;
    const isOAS3 = specSelectors.isOAS3();

    let parameterWithMeta = specSelectors.parameterWithMetaByIdentity(pathMethod, rawParam) || new Map();
    // fallback, if the meta lookup fails
    parameterWithMeta = parameterWithMeta.isEmpty() ? rawParam : parameterWithMeta;

    let enumValue;

    if (isOAS3) {
      const { schema } = getParameterSchema(parameterWithMeta, { isOAS3 });
      enumValue = schema ? schema.get('enum') : undefined;
    } else {
      enumValue = parameterWithMeta ? parameterWithMeta.get('enum') : undefined;
    }
    const paramValue = parameterWithMeta ? parameterWithMeta.get('value') : undefined;
    // input的值
    // console.log('paraValue', paramValue);
    // this.autoGenerateAppid(paramValue);
    let value;

    if (paramValue !== undefined) {
      value = paramValue;
    } else if (rawParam.get('required') && enumValue && enumValue.size) {
      value = enumValue.first();
    }

    if (value !== undefined && value !== paramValue) {
      this.onChangeWrapper(numberToString(value));
    }
    // todo: could check if schema here; if not, do not call. impact?
    this.setDefaultValue();
  }

  onChangeWrapper = (value, isXml = false) => {
    const { onChange, rawParam } = this.props;
    let valueForUpstream;

    // Coerce empty strings and empty Immutable objects to null
    if (value === '' || (value && value.size === 0)) {
      valueForUpstream = null;
    } else {
      valueForUpstream = value;
    }

    return onChange(rawParam, valueForUpstream, isXml);
  };

  _onExampleSelect = (key /* { isSyntheticChange } = {} */) => {
    this.props.oas3Actions.setActiveExamplesMember({
      name: key,
      pathMethod: this.props.pathMethod,
      contextType: 'parameters',
      contextName: this.getParamKey(),
    });
  };

  onChangeIncludeEmpty = newValue => {
    const { specActions, param, pathMethod } = this.props;
    const paramName = param.get('name');
    const paramIn = param.get('in');
    return specActions.updateEmptyParamInclusion(pathMethod, paramName, paramIn, newValue);
  };

  setDefaultValue = () => {
    const { specSelectors, pathMethod, rawParam, oas3Selectors } = this.props;

    const paramWithMeta = specSelectors.parameterWithMetaByIdentity(pathMethod, rawParam) || Map();
    const { schema } = getParameterSchema(paramWithMeta, { isOAS3: specSelectors.isOAS3() });
    const parameterMediaType = paramWithMeta.get('content', Map()).keySeq().first();

    // getSampleSchema could return null
    const generatedSampleValue = schema
      ? getSampleSchema(schema.toJS(), parameterMediaType, {
          includeWriteOnly: true,
        })
      : null;

    if (!paramWithMeta || paramWithMeta.get('value') !== undefined) {
      return;
    }

    if (paramWithMeta.get('in') !== 'body') {
      let initialValue;

      // Find an initial value

      if (specSelectors.isSwagger2()) {
        initialValue =
          paramWithMeta.get('x-example') !== undefined
            ? paramWithMeta.get('x-example')
            : paramWithMeta.getIn(['schema', 'example']) !== undefined
              ? paramWithMeta.getIn(['schema', 'example'])
              : schema && schema.getIn(['default']);
      } else if (specSelectors.isOAS3()) {
        const currentExampleKey = oas3Selectors.activeExamplesMember(...pathMethod, 'parameters', this.getParamKey());
        initialValue =
          paramWithMeta.getIn(['examples', currentExampleKey, 'value']) !== undefined
            ? paramWithMeta.getIn(['examples', currentExampleKey, 'value'])
            : paramWithMeta.getIn(['content', parameterMediaType, 'example']) !== undefined
              ? paramWithMeta.getIn(['content', parameterMediaType, 'example'])
              : paramWithMeta.get('example') !== undefined
                ? paramWithMeta.get('example')
                : (schema && schema.get('example')) !== undefined
                  ? schema && schema.get('example')
                  : (schema && schema.get('default')) !== undefined
                    ? schema && schema.get('default')
                    : paramWithMeta.get('default'); // ensures support for `parameterMacro`
      }

      // Process the initial value

      if (initialValue !== undefined && !List.isList(initialValue)) {
        // Stringify if it isn't a List
        initialValue = stringify(initialValue);
      }

      // Dispatch the initial value

      if (initialValue !== undefined) {
        this.onChangeWrapper(initialValue);
      } else if (schema && schema.get('type') === 'object' && generatedSampleValue && !paramWithMeta.get('examples')) {
        // Object parameters get special treatment.. if the user doesn't set any
        // default or example values, we'll provide initial values generated from
        // the schema.
        // However, if `examples` exist for the parameter, we won't do anything,
        // so that the appropriate `examples` logic can take over.
        this.onChangeWrapper(
          List.isList(generatedSampleValue) ? generatedSampleValue : stringify(generatedSampleValue)
        );
      }
    }
  };

  getParamKey() {
    const { param } = this.props;

    if (!param) return null;

    return `${param.get('name')}-${param.get('in')}`;
  }

  componentDidMount() {
    const disabledInputArr = ['appid'];
    document.addEventListener('click', () => {
      const input = document.getElementsByClassName('parameters-col_description');
      if (input.length > 0) {
        const inputArr = Array.from(input);
        inputArr.forEach(item => {
          const inputDOM = item.getElementsByTagName('input')[0];
          if (inputDOM && disabledInputArr.includes(inputDOM.placeholder?.toLowerCase())) {
            inputDOM.disabled = true;
          }
        });
      }
    });
  }

  render() {
    let {
      param,
      rawParam,
      getComponent,
      getConfigs,
      isExecute,
      fn,
      onChangeConsumes,
      specSelectors,
      pathMethod,
      specPath,
      oas3Selectors,
    } = this.props;

    const isOAS3 = specSelectors.isOAS3();

    const { showExtensions, showCommonExtensions } = getConfigs();

    if (!param) {
      param = rawParam;
    }

    if (!rawParam) return null;

    // const onChangeWrapper = (value) => onChange(param, value)
    // const JsonSchemaForm = getComponent('JsonSchemaForm');
    // const ParamBody = getComponent('ParamBody');
    const inType = param.get('in');
    const bodyParam =
      inType !== 'body' ? null : (
        <ParamBody
          getComponent={getComponent}
          getConfigs={getConfigs}
          fn={fn}
          param={param}
          consumes={specSelectors.consumesOptionsFor(pathMethod)}
          consumesValue={specSelectors.contentTypeValues(pathMethod).get('requestContentType')}
          onChange={this.onChangeWrapper}
          onChangeConsumes={onChangeConsumes}
          isExecute={isExecute}
          specSelectors={specSelectors}
          pathMethod={pathMethod}
        />
      );

    // const ModelExample = getComponent('modelExample');
    const Markdown = getComponent('Markdown', true);
    const ParameterExt = getComponent('ParameterExt');
    const ParameterIncludeEmpty = getComponent('ParameterIncludeEmpty');
    const ExamplesSelectValueRetainer = getComponent('ExamplesSelectValueRetainer');
    const Example = getComponent('Example');

    const { schema } = getParameterSchema(param, { isOAS3 });
    const paramWithMeta = specSelectors.parameterWithMetaByIdentity(pathMethod, rawParam) || Map();

    const format = schema ? schema.get('format') : null;
    const type = schema ? schema.get('type') : null;
    const itemType = schema ? schema.getIn(['items', 'type']) : null;
    const isFormData = inType === 'formData';
    const isFormDataSupported = 'FormData' in win;
    const required = param.get('required');

    const value = paramWithMeta ? paramWithMeta.get('value') : '';
    const commonExt = showCommonExtensions ? getCommonExtensions(schema) : null;
    const extensions = showExtensions ? getExtensions(param) : null;

    let paramItems; // undefined
    let paramEnum; // undefined
    let paramDefaultValue; // undefined
    let paramExample; // undefined
    let isDisplayParamEnum = false;

    if (param !== undefined && schema) {
      paramItems = schema.get('items');
    }

    if (paramItems !== undefined) {
      paramEnum = paramItems.get('enum');
      paramDefaultValue = paramItems.get('default');
    } else if (schema) {
      paramEnum = schema.get('enum');
    }

    if (paramEnum && paramEnum.size && paramEnum.size > 0) {
      isDisplayParamEnum = true;
    }

    // Default and Example Value for readonly doc
    if (param !== undefined) {
      if (schema) {
        paramDefaultValue = schema.get('default');
      }
      if (paramDefaultValue === undefined) {
        paramDefaultValue = param.get('default');
      }
      paramExample = param.get('example');
      if (paramExample === undefined) {
        paramExample = param.get('x-example');
      }
    }

    const paramName = param.get('name');

    // 当选中了【以用户token调试】，则Authorization对应的值要密文显示token，且不允许编辑。这里仅仅是显示，所以用了101位的*
    const paramValue = paramName === 'Authorization' && this.props.filter === FilterEnum.User ? '*'.repeat(101) : value;
    const paramDescription = param.get('description');

    const JsonSchemaFormDisabled =
      !isExecute || paramName === 'app_key' || (paramName === 'Authorization' && this.props.filter === FilterEnum.User);

    return (
      <tr data-param-name={paramName} data-param-in={param.get('in')}>
        <td className="parameters-col_name" style={{ width: '17%' }}>
          <div className={required ? 'parameter__name required' : 'parameter__name'}>
            {paramName}
            {!required ? null : <span>&nbsp;*</span>}
          </div>
          <div className="parameter__type">
            {type}
            {itemType && `[${itemType}]`}
            {format && <span className="prop-format">(${format})</span>}
          </div>
          <div className="parameter__deprecated">{isOAS3 && param.get('deprecated') ? 'deprecated' : null}</div>
          <div className="parameter__in">({param.get('in')})</div>
          {!showCommonExtensions || !commonExt.size
            ? null
            : commonExt.entrySeq().map(([key, v]) => <ParameterExt key={`${key}-${v}`} xKey={key} xVal={v} />)}
          {!showExtensions || !extensions.size
            ? null
            : extensions.entrySeq().map(([key, v]) => <ParameterExt key={`${key}-${v}`} xKey={key} xVal={v} />)}
        </td>

        <td className="parameters-col_description">
          {paramDescription ? (
            paramName === 'Authorization' && this.props.filter === FilterEnum.App ? (
              <div className="dip-mt-14 dip-mb-14">
                <span>{paramDescription}</span>
                <Tooltip title={'请联系管理员创建应用账号获取token'} className="dip-ml-6 dip-pointer">
                  <QuestionCircleOutlined />
                </Tooltip>
              </div>
            ) : (
              <Markdown source={paramDescription} />
            )
          ) : null}

          {(bodyParam || !isExecute) && isDisplayParamEnum ? (
            <Markdown
              className="parameter__enum"
              source={`<i>Available values</i> : ${paramEnum
                .map(item => {
                  return item;
                })
                .toArray()
                .join(', ')}`}
            />
          ) : null}

          {(bodyParam || !isExecute) && paramDefaultValue !== undefined ? (
            <Markdown className="parameter__default" source={'<i>Default value</i> : ' + paramDefaultValue} />
          ) : null}

          {(bodyParam || !isExecute) && paramExample !== undefined ? (
            <Markdown source={'<i>Example</i> : ' + paramExample} />
          ) : null}

          {isFormData && !isFormDataSupported && <div>Error: your browser does not support FormData</div>}

          {isOAS3 && param.get('examples') ? (
            <section className="parameter-controls">
              <ExamplesSelectValueRetainer
                examples={param.get('examples')}
                onSelect={this._onExampleSelect}
                updateValue={this.onChangeWrapper}
                getComponent={getComponent}
                defaultToFirstExample={true}
                currentKey={oas3Selectors.activeExamplesMember(...pathMethod, 'parameters', this.getParamKey())}
                currentUserInputValue={value}
              />
            </section>
          ) : null}

          {bodyParam ? null : (
            <JsonSchemaForm
              fn={fn}
              getComponent={getComponent}
              value={paramValue}
              required={required}
              disabled={JsonSchemaFormDisabled}
              description={paramName}
              onChange={this.onChangeWrapper}
              errors={paramWithMeta.get('errors')}
              schema={schema}
            />
          )}

          {bodyParam && schema ? (
            <ModelExample
              getComponent={getComponent}
              specPath={specPath.push('schema')}
              getConfigs={getConfigs}
              isExecute={isExecute}
              specSelectors={specSelectors}
              schema={schema}
              example={bodyParam}
              includeWriteOnly={true}
            />
          ) : null}

          {!bodyParam && isExecute && param.get('allowEmptyValue') ? (
            <ParameterIncludeEmpty
              onChange={this.onChangeIncludeEmpty}
              isIncluded={specSelectors.parameterInclusionSettingFor(pathMethod, paramName, param.get('in'))}
              isDisabled={!isEmptyValue(value)}
            />
          ) : null}

          {isOAS3 && param.get('examples') ? (
            <Example
              example={param.getIn([
                'examples',
                oas3Selectors.activeExamplesMember(...pathMethod, 'parameters', this.getParamKey()),
              ])}
              getComponent={getComponent}
              getConfigs={getConfigs}
            />
          ) : null}
        </td>
      </tr>
    );
  }
}
