/**
 * @prettier
 */

import React from 'react';
import Im from 'immutable';
import _ from 'lodash';
import PropTypes from 'prop-types';
import ImPropTypes from 'react-immutable-proptypes';
import DownSvg from '@/assets/icons/down-select.svg';

export default class ExamplesSelect extends React.PureComponent {
  state = { textContent: '' };
  static propTypes = {
    examples: ImPropTypes.map.isRequired,
    onSelect: PropTypes.func,
    currentExampleKey: PropTypes.string,
    isModifiedValueAvailable: PropTypes.bool,
    isValueModified: PropTypes.bool,
    showLabels: PropTypes.bool,
  };

  static defaultProps = {
    examples: Im.Map({}),
    onSelect: (...args) =>
      console.log(
        // FIXME: remove before merging to master...
        `DEBUG: ExamplesSelect was not given an onSelect callback`,
        ...args
      ),
    currentExampleKey: null,
    showLabels: true,
  };

  _onSelect = (key, { isSyntheticChange = false } = {}) => {
    if (typeof this.props.onSelect === 'function') {
      this.props.onSelect(key, {
        isSyntheticChange,
      });
    }
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

  _onDomSelect = e => {
    this._onChangeColor();
    if (typeof this.props.onSelect === 'function') {
      const element = e;
      // const element = e.target.selectedOptions[0];
      const key = element.getAttribute('value');

      this._onSelect(key, {
        isSyntheticChange: false,
      });
    }
  };

  getCurrentExample = () => {
    const { examples, currentExampleKey } = this.props;

    const currentExamplePerProps = examples.get(currentExampleKey);

    const firstExamplesKey = examples.keySeq().first();
    const firstExample = examples.get(firstExamplesKey);

    return currentExamplePerProps || firstExample || Map({});
  };

  _prevent = e => {
    if (e.stopPropagation)
      e.stopPropagation(); // 停止冒泡 非ie
    else e.cancelBubble = true;
  };

  componentDidMount() {
    // this is the not-so-great part of ExamplesSelect... here we're
    // artificially kicking off an onSelect event in order to set a default
    // value in state. the consumer has the option to avoid this by checking
    // `isSyntheticEvent`, but we should really be doing this in a selector.
    // TODO: clean this up
    // FIXME: should this only trigger if `currentExamplesKey` is nullish?

    const select = document.getElementsByClassName('select');
    _.map(select, t => {
      const selectHeadCont = t.getElementsByClassName('select-head-cont')[0];
      const optionItem = t.getElementsByClassName('option-item');
      const option = t.getElementsByClassName('option');
      selectHeadCont.innerHTML = optionItem[0].innerHTML;

      _.map(optionItem, item => {
        item.addEventListener('click', e => {
          selectHeadCont.innerHTML = item.innerHTML;
          this._onDomSelect(item);
          option[0].setAttribute('style', 'display:none;');
          this._prevent(e);
        });
      });

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
    });

    const { onSelect, examples } = this.props;

    if (typeof onSelect === 'function') {
      const firstExample = examples.first();
      const firstExampleKey = examples.keyOf(firstExample);

      this._onSelect(firstExampleKey, {
        isSyntheticChange: true,
      });
    }
  }

  UNSAFE_componentWillReceiveProps(nextProps) {
    const { currentExampleKey, examples } = nextProps;
    if (examples !== this.props.examples && !examples.has(currentExampleKey)) {
      // examples have changed from under us, and the currentExampleKey is no longer
      // valid.
      const firstExample = examples.first();
      const firstExampleKey = examples.keyOf(firstExample);

      this._onSelect(firstExampleKey, {
        isSyntheticChange: true,
      });
    }
  }

  render() {
    const { examples, currentExampleKey, isValueModified, isModifiedValueAvailable, showLabels } = this.props;

    return (
      <div className="examples-select">
        {showLabels ? <span className="examples-select__section-label">Examples: </span> : null}
        <ul className="select">
          <li style={{ position: 'relative' }}>
            <div className="select-head">
              <span className="select-head-cont">
                {/* {this.state.textContent} */}
                {/* {isModifiedValueAvailable && isValueModified ? '__MODIFIED__VALUE__' : currentExampleKey || ''} */}
              </span>
              <span className="select-icon">
                <DownSvg className="select-arrow" />
              </span>
            </div>
            <ul className="option">
              {isModifiedValueAvailable ? (
                <li className="option-item" value="__MODIFIED__VALUE__">
                  [Modified value]
                </li>
              ) : null}
              {examples
                .map((example, exampleName) => {
                  return (
                    <li className="option-item" key={exampleName} value={exampleName}>
                      {example.get('summary') || exampleName}
                    </li>
                  );
                })
                .valueSeq()}
            </ul>
          </li>
        </ul>
      </div>
    );
  }
}
