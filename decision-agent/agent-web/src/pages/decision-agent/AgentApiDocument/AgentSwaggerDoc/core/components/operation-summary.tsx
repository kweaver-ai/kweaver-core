import { PureComponent } from 'react';
import PropTypes from 'prop-types';
import _ from 'lodash';
import { Iterable, List } from 'immutable';
import intl from 'react-intl-universal';
import { message } from 'antd';
import ImPropTypes from 'react-immutable-proptypes';
import toString from 'lodash/toString';
import OperationSummaryMethod from './operation-summary-method';
import CopyToClipboardBtn from './copy-to-clipboard-btn';
import OperationSummaryPath from './operation-summary-path';

export default class OperationSummary extends PureComponent {
  static propTypes = {
    specPath: ImPropTypes.list.isRequired,
    operationProps: PropTypes.instanceOf(Iterable).isRequired,
    isShown: PropTypes.bool.isRequired,
    toggleShown: PropTypes.func.isRequired,
    getComponent: PropTypes.func.isRequired,
    getConfigs: PropTypes.func.isRequired,
    authActions: PropTypes.object,
    authSelectors: PropTypes.object,
  };

  static defaultProps = {
    operationProps: null,
    specPath: List(),
    summary: '',
  };

  componentDidMount() {
    this.props.toggleShown();
  }

  render() {
    const { isShown, toggleShown, getComponent, authActions, authSelectors, operationProps, specPath } = this.props;
    const {
      summary,
      isAuthorized,
      method,
      op,
      showSummary,
      path,
      operationId,
      originalOperationId,
      displayOperationId,
    } = operationProps.toJS();

    const { summary: resolvedSummary } = op;

    const security = operationProps.get('security');

    const AuthorizeOperationBtn = getComponent('authorizeOperationBtn');
    // const OperationSummaryMethod = getComponent('OperationSummaryMethod');
    // const OperationSummaryPath = getComponent('OperationSummaryPath');
    const JumpToPath = getComponent('JumpToPath', true);
    // const CopyToClipboardBtn = getComponent('CopyToClipboardBtn', true);

    const hasSecurity = security && !!security.count();
    const securityIsOptional = hasSecurity && security.size === 1 && security.first().isEmpty();
    const allowAnonymous = !hasSecurity || securityIsOptional;
    return (
      <div>
        <div
          className={`opblock-summary opblock-summary-${method}`}
          style={{ background: '#ffffff', border: '1px solid rgba(0,0,0,0.1)', width: '100%' }}
        >
          <button
            aria-label={`${method} ${path.replace(/\//g, '\u200b/')}`}
            aria-expanded={isShown}
            className="opblock-summary-control"
            // onClick={toggleShown}
            style={{ cursor: 'default', border: 'none', outline: 'none' }}
          >
            <OperationSummaryMethod method={method} />
            <OperationSummaryPath getComponent={getComponent} operationProps={operationProps} specPath={specPath} />

            {!showSummary ? null : (
              <div className="opblock-summary-description">{toString(resolvedSummary || summary)}</div>
            )}

            {displayOperationId && (originalOperationId || operationId) ? (
              <span className="opblock-summary-operation-id">{originalOperationId || operationId}</span>
            ) : null}
          </button>
          <CopyToClipboardBtn
            textToCopy={`${specPath?.get(1)}`}
            onClick={() => message.success(intl.get('dataAgent.copyKeySuccess'))}
          />
          <JumpToPath path={specPath} />
          {/* TODO: use wrapComponents here, swagger-ui doesn't care about jumpToPath */}
        </div>
      </div>
    );
  }
}
