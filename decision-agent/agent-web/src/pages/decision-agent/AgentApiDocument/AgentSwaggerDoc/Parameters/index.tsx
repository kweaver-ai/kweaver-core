import { Component } from 'react';
import PropTypes from 'prop-types';
import { List, Map } from 'immutable';
import ImPropTypes from 'react-immutable-proptypes';
import { getComponentEventEmitter, ComponentIDEnum } from '@/utils/event-bus';
import { FilterEnum } from '../types';
import TryItOutButton from '../core/components/try-it-out-button';
import RequestBody from '../core/plugins/oas3/components/request-body';
import ParameterRow from '../core/components/parameter-row';
import Callbacks from '../core/plugins/oas3/components/callbacks';
import ContentType from '../core/components/content-type';

export default class Parameters extends Component<any> {
  state = {
    callbackVisible: false,
    parametersVisible: true,
    filter: FilterEnum.User,
  };

  static propTypes = {
    parameters: ImPropTypes.list.isRequired,
    operation: PropTypes.object.isRequired,
    specActions: PropTypes.object.isRequired,
    getComponent: PropTypes.func.isRequired,
    specSelectors: PropTypes.object.isRequired,
    oas3Actions: PropTypes.object.isRequired,
    oas3Selectors: PropTypes.object.isRequired,
    fn: PropTypes.object.isRequired,
    tryItOutEnabled: PropTypes.bool,
    allowTryItOut: PropTypes.bool,
    onTryoutClick: PropTypes.func,
    onResetClick: PropTypes.func,
    onCancelClick: PropTypes.func,
    onChangeKey: PropTypes.array,
    pathMethod: PropTypes.array.isRequired,
    getConfigs: PropTypes.func.isRequired,
    specPath: ImPropTypes.list.isRequired,
  };

  static defaultProps = {
    onTryoutClick: Function.prototype,
    onCancelClick: Function.prototype,
    tryItOutEnabled: false,
    allowTryItOut: true,
    onChangeKey: [],
    specPath: [],
  };

  emitter = getComponentEventEmitter(ComponentIDEnum.APIDoc);

  componentDidMount(): void {
    // 监听filter变化，更新value
    this.emitter.on('changeFilter', this.changeValue);
  }

  componentWillUnmount(): void {
    this.emitter.off('changeFilter', this.changeValue);
  }

  changeValue = ({ filter }: { filter: FilterEnum }) => {
    this.setState({
      filter,
    });
  };

  onChange = (param: any, value: any, isXml: any) => {
    const {
      specActions: { changeParamByIdentity },
      onChangeKey,
    } = this.props;

    changeParamByIdentity(onChangeKey, param, value, isXml);
  };

  onChangeConsumesWrapper = (val: any) => {
    const {
      specActions: { changeConsumesValue },
      onChangeKey,
    } = this.props;

    changeConsumesValue(onChangeKey, val);
  };
  toggleTab = (tab: string) => {
    if (tab === 'parameters') {
      return this.setState({
        parametersVisible: true,
        callbackVisible: false,
      });
    } else if (tab === 'callbacks') {
      return this.setState({
        callbackVisible: true,
        parametersVisible: false,
      });
    }
  };

  onChangeMediaType = ({ value, pathMethod }: any) => {
    const { specActions, oas3Selectors, oas3Actions } = this.props;
    const userHasEditedBody = oas3Selectors.hasUserEditedBody(...pathMethod);
    const shouldRetainRequestBodyValue = oas3Selectors.shouldRetainRequestBodyValue(...pathMethod);
    oas3Actions.setRequestContentType({ value, pathMethod });
    oas3Actions.initRequestBodyValidateError({ pathMethod });
    if (!userHasEditedBody) {
      if (!shouldRetainRequestBodyValue) {
        oas3Actions.setRequestBodyValue({ value: undefined, pathMethod });
      }
      specActions.clearResponse(...pathMethod);
      specActions.clearRequest(...pathMethod);
      specActions.clearValidateParams(pathMethod);
    }
  };

  render() {
    const {
      onTryoutClick,
      onResetClick,
      parameters,
      allowTryItOut,
      tryItOutEnabled,
      specPath,
      fn,
      getComponent,
      getConfigs,
      specSelectors,
      specActions,
      pathMethod,
      oas3Actions,
      oas3Selectors,
      operation,
    } = this.props;

    // const ParameterRow = getComponent('parameterRow');
    // const TryItOutButton = getComponent('TryItOutButton');
    // const ContentType = getComponent('contentType');
    // const Callbacks = getComponent('Callbacks', true);
    // const RequestBody = getComponent('RequestBody', true);

    const isExecute = tryItOutEnabled && allowTryItOut;
    const isOAS3 = specSelectors.isOAS3();

    const requestBody = operation.get('requestBody');

    const groupedParametersArr = Object.values(
      parameters.reduce((acc: { [x: string]: any[] }, x: { get: (arg0: string) => any }) => {
        const key = x.get('in');
        !acc[key] && (acc[key] = []);
        acc[key].push(x);
        return acc;
      }, {})
    ).reduce((acc: any, x) => acc.concat(x), []);

    const retainRequestBodyValueFlagForOperation = (f: any) =>
      oas3Actions.setRetainRequestBodyValueFlag({ value: f, pathMethod });
    return (
      <div className="opblock-section">
        <div className="opblock-section-header" style={{ padding: '0px', boxShadow: 'none' }}>
          {isOAS3 ? (
            <div className="tab-header">
              <div
                onClick={() => this.toggleTab('parameters')}
                className={`tab-item ${this.state.parametersVisible && 'active'}`}
              >
                <h4 className="opblock-title">
                  <span>Parameters</span>
                </h4>
              </div>
              {operation.get('callbacks') ? (
                <div
                  onClick={() => this.toggleTab('callbacks')}
                  className={`tab-item ${this.state.callbackVisible && 'active'}`}
                >
                  <h4 className="opblock-title">
                    <span>Callbacks</span>
                  </h4>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="tab-header">
              <h4 className="opblock-title">Parameters</h4>
            </div>
          )}
          {allowTryItOut ? (
            <TryItOutButton
              isOAS3={specSelectors.isOAS3()}
              hasUserEditedBody={oas3Selectors.hasUserEditedBody(...pathMethod)}
              enabled={tryItOutEnabled}
              onCancelClick={this.props.onCancelClick}
              onTryoutClick={onTryoutClick}
              onResetClick={() => onResetClick(pathMethod)}
            />
          ) : null}
        </div>
        {this.state.parametersVisible ? (
          <div className="parameters-container">
            {!(groupedParametersArr as any).length ? (
              <div className="opblock-description-wrapper">
                <p>No parameters</p>
              </div>
            ) : (
              <div className="table-container">
                <table className="parameters">
                  <thead>
                    <tr>
                      {
                        <>
                          <th className="col_header parameters-col_name">Name</th>
                          <th className="col_header parameters-col_description">Description</th>
                        </>
                      }
                    </tr>
                  </thead>
                  <tbody>
                    {(groupedParametersArr as any).map(
                      (parameter: { get: (arg0: string) => any }, i: { toString: () => any }) => (
                        <ParameterRow
                          fn={fn}
                          specPath={specPath.push(i.toString())}
                          getComponent={getComponent}
                          getConfigs={getConfigs}
                          rawParam={parameter}
                          param={specSelectors.parameterWithMetaByIdentity(pathMethod, parameter)}
                          key={`${parameter.get('in')}.${parameter.get('name')}`}
                          onChange={this.onChange}
                          onChangeConsumes={this.onChangeConsumesWrapper}
                          specSelectors={specSelectors}
                          specActions={specActions}
                          oas3Actions={oas3Actions}
                          oas3Selectors={oas3Selectors}
                          pathMethod={pathMethod}
                          isExecute={isExecute}
                          filter={this.state.filter}
                        />
                      )
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        ) : null}

        {this.state.callbackVisible ? (
          <div className="callbacks-container opblock-description-wrapper">
            <Callbacks
              getComponent={getComponent}
              callbacks={Map(operation.get('callbacks'))}
              specPath={specPath.slice(0, -1).push('callbacks')}
            />
          </div>
        ) : null}
        {isOAS3 && requestBody && this.state.parametersVisible && (
          <div className="opblock-section opblock-section-request-body">
            <div className="opblock-section-header" style={{ padding: '0px', boxShadow: 'none' }}>
              <h4 className={`opblock-title parameter__name ${requestBody.get('required') && 'required'}`}>
                Request body
              </h4>
              <label>
                <ContentType
                  value={oas3Selectors.requestContentType(...pathMethod)}
                  contentTypes={requestBody.get('content', List()).keySeq()}
                  onChange={(value: any) => {
                    this.onChangeMediaType({ value, pathMethod });
                  }}
                  className="body-param-content-type"
                  ariaLabel="Request content type"
                />
              </label>
            </div>
            <div className="opblock-description-wrapper">
              <RequestBody
                setRetainRequestBodyValueFlag={retainRequestBodyValueFlagForOperation}
                userHasEditedBody={oas3Selectors.hasUserEditedBody(...pathMethod)}
                specPath={specPath.slice(0, -1).push('requestBody')}
                requestBody={requestBody}
                requestBodyValue={oas3Selectors.requestBodyValue(...pathMethod)}
                requestBodyInclusionSetting={oas3Selectors.requestBodyInclusionSetting(...pathMethod)}
                requestBodyErrors={oas3Selectors.requestBodyErrors(...pathMethod)}
                isExecute={isExecute}
                fn={fn}
                oas3Actions={oas3Actions}
                specSelectors={specSelectors}
                getComponent={getComponent}
                getConfigs={getConfigs}
                activeExamplesKey={oas3Selectors.activeExamplesMember(
                  ...pathMethod,
                  'requestBody',
                  'requestBody' // RBs are currently not stored per-mediaType
                )}
                updateActiveExamplesKey={(key: any) => {
                  this.props.oas3Actions.setActiveExamplesMember({
                    name: key,
                    pathMethod: this.props.pathMethod,
                    contextType: 'requestBody',
                    contextName: 'requestBody', // RBs are currently not stored per-mediaType
                  });
                }}
                onChange={(value: any, path: Iterable<any>) => {
                  if (path) {
                    const lastValue = oas3Selectors.requestBodyValue(...pathMethod);
                    const usableValue = Map.isMap(lastValue) ? lastValue : Map();
                    return oas3Actions.setRequestBodyValue({
                      pathMethod,
                      value: usableValue.setIn(path, value),
                    });
                  }
                  oas3Actions.setRequestBodyValue({ value, pathMethod });
                }}
                onChangeIncludeEmpty={(name: any, value: any) => {
                  oas3Actions.setRequestBodyInclusion({
                    pathMethod,
                    value,
                    name,
                  });
                }}
                contentType={oas3Selectors.requestContentType(...pathMethod)}
              />
            </div>
          </div>
        )}
      </div>
    );
  }
}
