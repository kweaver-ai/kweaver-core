import React from 'react';
import PropTypes from 'prop-types';
import intl from 'react-intl-universal';
import { getConfig } from '@/utils/http';

export default class TryItOutButton extends React.Component {
  state = {
    appVisible: false,
  };
  static propTypes = {
    onTryoutClick: PropTypes.func,
    onResetClick: PropTypes.func,
    onCancelClick: PropTypes.func,
    enabled: PropTypes.bool, // Try it out is enabled, ie: the user has access to the form
    hasUserEditedBody: PropTypes.bool, // Try it out is enabled, ie: the user has access to the form
    isOAS3: PropTypes.bool, // Try it out is enabled, ie: the user has access to the form
  };

  static defaultProps = {
    onTryoutClick: Function.prototype,
    onCancelClick: Function.prototype,
    onResetClick: Function.prototype,
    enabled: false,
    hasUserEditedBody: false,
    isOAS3: false,
  };

  onAcquire = () => {
    this.setState({ appVisible: true });
  };

  render() {
    const { onTryoutClick, onCancelClick, onResetClick, enabled, hasUserEditedBody, isOAS3 } = this.props;

    const showReset = isOAS3 && hasUserEditedBody;
    return (
      <div className={showReset ? 'try-out btn-group' : 'try-out'}>
        {enabled ? (
          <button className="btn try-out__btn cancel" type="default" onClick={onCancelClick}>
            {intl.get('dataAgent.cancel')}
          </button>
        ) : (
          <button
            className="btn try-out__btn"
            style={{ width: '120px', background: getConfig('theme'), color: '#ffffff', border: 'none' }}
            onClick={onTryoutClick}
          >
            {intl.get('dataAgent.apiDocument.debug')}{' '}
          </button>
        )}
        {showReset && (
          <button className="btn try-out__btn reset" type="default" style={{ width: '120px' }} onClick={onResetClick}>
            {intl.get('dataAgent.apiDocument.reset')}
          </button>
        )}
      </div>
    );
  }
}
