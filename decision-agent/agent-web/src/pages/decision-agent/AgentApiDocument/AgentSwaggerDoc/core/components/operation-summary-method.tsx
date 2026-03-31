import { PureComponent } from 'react';
import PropTypes from 'prop-types';
import { Iterable } from 'immutable';

export default class OperationSummaryMethod extends PureComponent {
  static propTypes = {
    operationProps: PropTypes.instanceOf(Iterable).isRequired,
    method: PropTypes.string.isRequired,
  };

  static defaultProps = {
    operationProps: null,
  };
  render() {
    const { method } = this.props;

    return (
      <span style={{ background: '#ffffff', color: 'rgba(1,150,136,0.85)' }} className="opblock-summary-method">
        {method.toUpperCase()}
      </span>
    );
  }
}
