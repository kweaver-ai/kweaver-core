import React from 'react';
import PropTypes from 'prop-types';
import ImPropTypes from 'react-immutable-proptypes';
import { fromJS } from 'immutable';
import _ from 'lodash';
import DownSvg from '@/assets/icons/down-select.svg';

const noop = () => {};

export default class ContentType extends React.Component {
  static propTypes = {
    ariaControls: PropTypes.string,
    contentTypes: PropTypes.oneOfType([ImPropTypes.list, ImPropTypes.set, ImPropTypes.seq]),
    controlId: PropTypes.string,
    value: PropTypes.string,
    onChange: PropTypes.func,
    className: PropTypes.string,
    ariaLabel: PropTypes.string,
  };

  static defaultProps = {
    onChange: noop,
    value: null,
    contentTypes: fromJS(['application/json']),
  };

  _onChangeColor = () => {
    const selectResponse = document.getElementsByClassName('example microlight');
    _.map(selectResponse, item => {
      item.setAttribute('style', 'background:rgba(0,0,0,0.04);color:rgba(0,0,0,0.85);border:1px solid rgba(0,0,0,0.1)');
      _.map(item.getElementsByTagName('span'), i => {
        i.setAttribute('style', 'color:rgba(0,0,0,0.85)');
      });
    });
  };

  _prevent = e => {
    if (e.stopPropagation)
      e.stopPropagation(); // 停止冒泡 非ie
    else e.cancelBubble = true;
  };

  componentDidMount() {
    const content = document.getElementsByClassName('content-select');
    _.map(content, t => {
      const contentHeadCont = t.getElementsByClassName('content-head-cont')?.[0];
      const optionTypeItem = t.getElementsByClassName('option-type-item');
      const option = t.getElementsByClassName('option-type');
      contentHeadCont.innerHTML = optionTypeItem[0].innerHTML;
      t.addEventListener('click', e => {
        option[0].setAttribute(
          'style',
          'display:block;border:1px solid rgba(0,0,0,0.1);margin-top:6px;z-index:1000;position:absolute;background:white;width:200px;'
        );
        this._prevent(e);
      });

      document.addEventListener('click', () => {
        option[0].setAttribute('style', 'display:none;');
      });
      _.map(optionTypeItem, item => {
        item.addEventListener('click', e => {
          contentHeadCont.innerHTML = item.innerHTML;
          this.onChangeWrapper(item.innerHTML);
          option[0].setAttribute('style', 'display:none;');
          this._prevent(e);
        });
      });
    });

    // Needed to populate the form, initially
    if (this.props.contentTypes) {
      this.props.onChange(this.props.contentTypes.first());
    }
  }

  UNSAFE_componentWillReceiveProps(nextProps) {
    if (!nextProps.contentTypes || !nextProps.contentTypes.size) {
      return;
    }

    if (!nextProps.contentTypes.includes(nextProps.value)) {
      nextProps.onChange(nextProps.contentTypes.first());
    }
  }

  // onChangeWrapper = e => this.props.onChange(e.target.value);
  onChangeWrapper = e => {
    this._onChangeColor();
    this.props.onChange(e);
  };

  render() {
    const { ariaControls, ariaLabel, className, contentTypes, controlId, value } = this.props;

    if (!contentTypes || !contentTypes.size) return null;

    return (
      <div className={`content-type-wrapper ${className || ''}`}>
        <ul className="content-select" aria-controls={ariaControls} aria-label={ariaLabel}>
          <li style={{ position: 'relative' }}>
            <div className="content-head">
              <span className="content-head-cont"></span>
              <span className="content-icon">
                <DownSvg className="select-arrow" />
              </span>
            </div>
            <ul className="option-type">
              {contentTypes
                .map(val => {
                  return (
                    <li className="option-type-item" key={val} value={val}>
                      {val}
                    </li>
                  );
                })
                .toArray()}
            </ul>
          </li>
        </ul>
      </div>
    );
  }
}
