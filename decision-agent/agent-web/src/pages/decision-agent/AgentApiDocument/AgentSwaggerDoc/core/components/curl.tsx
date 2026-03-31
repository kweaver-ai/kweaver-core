import React from 'react';
import PropTypes from 'prop-types';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import intl from 'react-intl-universal';
import { message } from 'antd';
import get from 'lodash/get';
import { SyntaxHighlighter, getStyle } from '../syntax-highlighting';
import { requestSnippetGenerator_curl_bash } from '../plugins/request-snippets/fn';

export default class Curl extends React.Component {
  static propTypes = {
    getConfigs: PropTypes.func.isRequired,
    request: PropTypes.object.isRequired,
  };

  render() {
    const { request, getConfigs } = this.props;
    const curl = requestSnippetGenerator_curl_bash(request);

    const config = getConfigs();
    const styleCurl = getStyle(get(config, 'syntaxHighlight.theme'));
    styleCurl.hljs.background = 'rgba(0,0,0,0.04)';
    styleCurl.hljs.color = 'rgba(0,0,0,0.85)';
    styleCurl.hljs.padding = '0.5em 0.5em 12px';
    styleCurl.hljs.border = '1px solid rgba(0,0,0,0.1)';
    styleCurl['hljs-string'].color = 'rgba(0,0,0,0.85)';

    const curlBlock = get(config, 'syntaxHighlight.activated') ? (
      <SyntaxHighlighter language="bash" className="curl microlight" style={styleCurl}>
        {curl}
      </SyntaxHighlighter>
    ) : (
      <textarea readOnly={true} className="curl" value={curl}></textarea>
    );

    return (
      <div className="curl-command">
        <h4 className="dip-font-weight-700">Curl</h4>
        <div className="copy-to-clipboard">
          <CopyToClipboard text={curl}>
            <button onClick={() => message.success(intl.get('dataAgent.copyKeySuccess'))} />
          </CopyToClipboard>
        </div>
        <div>{curlBlock}</div>
      </div>
    );
  }
}
