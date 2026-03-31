import React from 'react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import PropTypes from 'prop-types';
import intl from 'react-intl-universal';
import { message } from 'antd';

/**
 * @param {{ textToCopy: string }} props
 * @returns {JSX.Element}
 * @constructor
 */
export default class CopyToClipboardBtn extends React.Component {
  onToast = () => {
    message.success(intl.get('dataAgent.copyKeySuccess'));
  };
  render() {
    const { origin } = window.location;
    return (
      <div
        className="view-line-link copy-to-clipboard"
        title="Copy to clipboard"
        style={{ width: '40px', position: 'position' }}
        onClick={() => this.onToast()}
      >
        <CopyToClipboard text={origin + this.props.textToCopy}>
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
    );
  }

  static propTypes = {
    textToCopy: PropTypes.string.isRequired,
  };
}
