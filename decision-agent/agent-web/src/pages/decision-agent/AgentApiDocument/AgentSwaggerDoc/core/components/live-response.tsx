import React from 'react';
import PropTypes from 'prop-types';
import ImPropTypes from 'react-immutable-proptypes';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import intl from 'react-intl-universal';
import { message } from 'antd';
import Curl from './curl';
import Markdown from './providers/markdown';
import ResponseBody from './response-body';
const judgement = {
  2: { background: 'rgba(82, 196, 26, 0.1)', color: '#52C41A' },
  3: { background: 'rgba(245, 34, 45, 0.1)', color: '#F5222D' },
  4: { background: 'rgba(245, 34, 45, 0.1)', color: '#F5222D' },
  5: { background: 'rgba(245, 34, 45, 0.1)', color: '#F5222D' },
};

const Headers = ({ headers }) => {
  return (
    <div>
      <h5 className="dip-font-weight-700">Response headers</h5>
      <pre
        className="microlight"
        style={{ background: 'rgba(0,0,0,0.04)', border: '1px solid rgba(0,0,0,0.1)', color: 'rgba(0,0,0,0.85)' }}
      >
        {headers}
      </pre>
    </div>
  );
};
Headers.propTypes = {
  headers: PropTypes.array.isRequired,
};

const Duration = ({ duration }) => {
  return (
    <div>
      <h5>Request duration</h5>
      <pre className="microlight">{duration} ms</pre>
    </div>
  );
};
Duration.propTypes = {
  duration: PropTypes.number.isRequired,
};

export default class LiveResponse extends React.Component {
  static propTypes = {
    response: ImPropTypes.map,
    path: PropTypes.string.isRequired,
    method: PropTypes.string.isRequired,
    displayRequestDuration: PropTypes.bool.isRequired,
    specSelectors: PropTypes.object.isRequired,
    getComponent: PropTypes.func.isRequired,
    getConfigs: PropTypes.func.isRequired,
  };

  shouldComponentUpdate(nextProps) {
    // BUG: props.response is always coming back as a new Immutable instance
    // same issue as responses.jsx (tryItOutResponse)
    return (
      this.props.response !== nextProps.response ||
      this.props.path !== nextProps.path ||
      this.props.method !== nextProps.method ||
      this.props.displayRequestDuration !== nextProps.displayRequestDuration
    );
  }

  render() {
    const { response, getComponent, getConfigs, displayRequestDuration, specSelectors, path, method } = this.props;
    const { showMutatedRequest, requestSnippetsEnabled } = getConfigs();

    const curlRequest = showMutatedRequest
      ? specSelectors.mutatedRequestFor(path, method)
      : specSelectors.requestFor(path, method);
    const status = response.get('status');
    const url = curlRequest?.get('url');
    const headers = response.get('headers').toJS();
    const notDocumented = response.get('notDocumented');
    const isError = response.get('error');
    const body = response.get('text');
    const duration = response.get('duration');
    const headersKeys = Object.keys(headers);
    const contentType = headers['content-type'] || headers['Content-Type'];

    // const ResponseBody = getComponent('responseBody');
    const returnObject = headersKeys.map(key => {
      const joinedHeaders = Array.isArray(headers[key]) ? headers[key].join() : headers[key];
      return (
        <span className="headerline" key={key}>
          {' '}
          {key}: {joinedHeaders}{' '}
        </span>
      );
    });
    const hasHeaders = returnObject.length !== 0;
    // const Markdown = getComponent('Markdown', true);
    const RequestSnippets = getComponent('RequestSnippets', true);
    // const Curl = getComponent('curl');

    return (
      <div>
        {curlRequest &&
          (requestSnippetsEnabled === true || requestSnippetsEnabled === 'true' ? (
            <RequestSnippets request={curlRequest} />
          ) : (
            <Curl request={curlRequest} getConfigs={getConfigs} />
          ))}
        {url && (
          <div>
            <div className="request-url">
              <h4 className="dip-font-weight-700">Request URL</h4>
              <pre
                className="microlight"
                style={{
                  background: 'rgba(0,0,0,0.04)',
                  border: '1px solid rgba(0,0,0,0.1)',
                  color: 'rgba(0,0,0,0.85)',
                  display: 'flex',
                  position: 'relative',
                }}
              >
                {url}
                <div className="copyBoard" onClick={() => message.success(intl.get('dataAgent.copyKeySuccess'))}>
                  <CopyToClipboard text={url}>
                    <svg
                      t="1677476982456"
                      viewBox="0 0 1024 1024"
                      version="1.1"
                      xmlns="http://www.w3.org/2000/svg"
                      p-id="15074"
                      width="15"
                      height="16"
                    >
                      <path
                        d="M600.576 267.264H240.128c-8.704 0-15.36 6.656-15.36 15.36v30.72c0 8.704 6.656 15.36 15.36 15.36h359.936c8.704 0 15.36-6.656 15.36-15.36v-30.72c0.512-8.704-6.656-15.36-14.848-15.36z m-87.04 243.712H240.128c-8.704 0-15.36 6.656-15.36 15.36v30.72c0 8.704 6.656 15.36 15.36 15.36h272.896c8.704 0 15.36-6.656 15.36-15.36v-30.72c0.512-8.192-6.656-15.36-14.848-15.36z"
                        p-id="15075"
                      ></path>
                      <path
                        d="M846.848 180.736h-73.216V138.24c0-67.584-55.296-122.88-122.88-122.88H178.176c-67.584 0-122.88 55.296-122.88 122.88v581.632c0 67.584 55.296 122.88 122.88 122.88h73.216v42.496c0 67.584 55.296 122.88 122.88 122.88h472.576c67.584 0 122.88-55.296 122.88-122.88V303.616c0-67.584-54.784-122.88-122.88-122.88zM116.736 720.384V138.24c0-33.792 27.648-61.44 61.44-61.44h472.576c33.792 0 61.44 27.648 61.44 61.44v581.632c0 33.792-27.648 61.44-61.44 61.44H178.176c-33.792 0.512-61.44-27.136-61.44-60.928zM908.288 885.76c0 33.792-27.648 61.44-61.44 61.44H374.272c-33.792 0-61.44-27.648-61.44-61.44v-42.496h338.432c67.584 0 122.88-55.296 122.88-122.88V242.176h73.216c33.792 0 61.44 27.648 61.44 61.44V885.76h-0.512z"
                        p-id="15076"
                      ></path>
                    </svg>
                  </CopyToClipboard>
                </div>
              </pre>
            </div>
          </div>
        )}
        <h4 className="server-h4">Server response</h4>
        {/* <di></di> */}
        <table className="responses-table live-responses-table">
          <thead>
            {/* <tr className="responses-header">
              {/* <td className="col_header response-col_status">Code</td>
              <td className="col_header response-col_description">Details</td> */}
            {/* </tr> */}
          </thead>
          <tbody>
            <tr className="response">
              {/* <td className="response-col_status">
                {status}
                {notDocumented ? (
                  <div className="response-undocumented">
                    <i> Undocumented </i>
                  </div>
                ) : null}
              </td> */}
              <td className="response-col_description">
                {/* {isError ? ( */}
                <div
                  className="dip-flex-align-center"
                  style={{
                    paddingLeft: '16px',
                    borderRadius: '4px',
                    height: '40px',
                    width: '100%',
                    ...judgement[String(status)[0]],
                  }}
                >
                  {status === 200 ? (
                    <>
                      <span className="dip-mr-16">{status}</span>
                      <span>OK</span>{' '}
                    </>
                  ) : (
                    <span className="dip-mr-16">{status}</span>
                  )}
                  {status !== 200 && (
                    <Markdown
                      source={`${response.get('name') !== '' ? `${response.get('name')}: ` : ''}${response.get(
                        'message'
                      )}`}
                    />
                  )}
                </div>
                {/* ) : null} */}
                {body ? (
                  <ResponseBody
                    content={body}
                    contentType={contentType}
                    url={url}
                    headers={headers}
                    getConfigs={getConfigs}
                    getComponent={getComponent}
                  />
                ) : null}
                {hasHeaders ? <Headers headers={returnObject} /> : null}
                {displayRequestDuration && duration ? <Duration duration={duration} /> : null}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    );
  }
}
