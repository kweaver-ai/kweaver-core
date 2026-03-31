import { Component } from 'react';
import PropTypes from 'prop-types';
import intl from 'react-intl-universal';

export default class Clear extends Component {
  onClick = () => {
    const { specActions, path, method } = this.props;
    specActions.clearResponse(path, method);
    specActions.clearRequest(path, method);
  };

  render() {
    return (
      <button className="btn btn-clear opblock-control__btn dip-ml-12" onClick={this.onClick}>
        {intl.get('dataAgent.apiDocument.clear')}
      </button>
    );
  }

  static propTypes = {
    specActions: PropTypes.object.isRequired,
    path: PropTypes.string.isRequired,
    method: PropTypes.string.isRequired,
  };
}
