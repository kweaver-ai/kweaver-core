import React, { PureComponent } from 'react';
import PropTypes from 'prop-types';
import { fromJS, List } from 'immutable';
import { getSampleSchema } from '../utils';
import { getKnownSyntaxHighlighterLanguage } from '../utils/jsonParse';

const NOOP = Function.prototype;

export default class ParamBody extends PureComponent {
  static propTypes = {
    param: PropTypes.object,
    onChange: PropTypes.func,
    onChangeConsumes: PropTypes.func,
    consumes: PropTypes.object,
    consumesValue: PropTypes.string,
    fn: PropTypes.object.isRequired,
    getConfigs: PropTypes.func.isRequired,
    getComponent: PropTypes.func.isRequired,
    isExecute: PropTypes.bool,
    specSelectors: PropTypes.object.isRequired,
    pathMethod: PropTypes.array.isRequired,
  };

  static defaultProp = {
    consumes: fromJS(['application/json']),
    param: fromJS({}),
    onChange: NOOP,
    onChangeConsumes: NOOP,
  };

  constructor(props, context) {
    super(props, context);

    this.state = {
      isEditBox: false,
      value: '',
    };
  }

  componentDidMount() {
    this.updateValues.call(this, this.props);
  }

  UNSAFE_componentWillReceiveProps(nextProps) {
    this.updateValues.call(this, nextProps);
  }

  updateValues = props => {
    const { param, isExecute, consumesValue = '' } = props;
    const isXml = /xml/i.test(consumesValue);
    const isJson = /json/i.test(consumesValue);
    const paramValue = isXml ? param.get('value_xml') : param.get('value');

    if (paramValue !== undefined) {
      const val = !paramValue && isJson ? '{}' : paramValue;
      this.setState({ value: val });
      this.onChange(val, { isXml: isXml, isEditBox: isExecute });
    } else {
      if (isXml) {
        this.onChange(this.sample('xml'), { isXml: isXml, isEditBox: isExecute });
      } else {
        this.onChange(this.sample(), { isEditBox: isExecute });
      }
    }
  };

  sample = xml => {
    const {
      param,
      fn: { inferSchema },
    } = this.props;
    const schema = inferSchema(param.toJS());

    return getSampleSchema(schema, xml, {
      includeWriteOnly: true,
    });
  };

  onChange = (value, { isEditBox, isXml }) => {
    this.setState({ value, isEditBox });
    this._onChange(value, isXml);
  };

  _onChange = (val, isXml) => {
    (this.props.onChange || NOOP)(val, isXml);
  };

  handleOnChange = e => {
    const { consumesValue } = this.props;
    const isXml = /xml/i.test(consumesValue);
    const inputValue = e.target.value;
    this.onChange(inputValue, { isXml, isEditBox: this.state.isEditBox });
  };

  toggleIsEditBox = () => this.setState(state => ({ isEditBox: !state.isEditBox }));

  render() {
    const { onChangeConsumes, param, isExecute, specSelectors, pathMethod, getConfigs, getComponent } = this.props;
    const Button = getComponent('Button');
    const TextArea = getComponent('TextArea');
    const HighlightCode = getComponent('highlightCode');
    const ContentType = getComponent('contentType');
    // for domains where specSelectors not passed
    const parameter = specSelectors ? specSelectors.parameterWithMetaByIdentity(pathMethod, param) : param;
    const errors = parameter.get('errors', List());
    const consumesValue = specSelectors.contentTypeValues(pathMethod).get('requestContentType');
    const consumes =
      this.props.consumes && this.props.consumes.size ? this.props.consumes : ParamBody.defaultProp.consumes;

    const { value, isEditBox } = this.state;
    let language = null;
    const testValueForJson = getKnownSyntaxHighlighterLanguage(value);
    if (testValueForJson) {
      language = 'json';
    }

    return (
      <div className="body-param" data-param-name={param.get('name')} data-param-in={param.get('in')}>
        {isEditBox && isExecute ? (
          <TextArea
            className={`body-param__text${errors.count() ? ' invalid' : ''}`}
            value={value}
            onChange={this.handleOnChange}
          />
        ) : (
          value && (
            <HighlightCode className="body-param__example" language={language} getConfigs={getConfigs} value={value} />
          )
        )}
        <div className="body-param-options">
          {!isExecute ? null : (
            <div className="body-param-edit">
              <Button
                className={isEditBox ? 'btn cancel body-param__example-edit' : 'btn edit body-param__example-edit'}
                onClick={this.toggleIsEditBox}
              >
                {isEditBox ? 'Cancel' : 'Edit'}
              </Button>
            </div>
          )}
          <label htmlFor="">
            <span>Parameter content type</span>
            <ContentType
              value={consumesValue}
              contentTypes={consumes}
              onChange={onChangeConsumes}
              className="body-param-content-type"
              ariaLabel="Parameter content type"
            />
          </label>
        </div>
      </div>
    );
  }
}
