import { useRef, useEffect } from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';
import intl from 'react-intl-universal';
import { message } from 'antd';
import get from 'lodash/get';
import isFunction from 'lodash/isFunction';
import saveAs from 'js-file-download';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import DownSvg from '@/assets/icons/down.svg';
import { SyntaxHighlighter, getStyle } from '../syntax-highlighting';

const HighlightCode = props => {
  const { value, fileName, className, downloadable, getConfigs, canCopy, language } = props;
  const config = isFunction(getConfigs) ? getConfigs() : null;
  const canSyntaxHighlight = get(config, 'syntaxHighlight') !== false && get(config, 'syntaxHighlight.activated', true);
  const rootRef = useRef(null);

  useEffect(() => {
    const syntaxStyle = getStyle(get(config, 'syntaxHighlight.theme', 'agate'));
    syntaxStyle.hljs.background = '#f5f5f5';
    syntaxStyle.hljs.color = 'rgba(0, 0, 0, 0.85)';
    syntaxStyle['hljs-string'].color = 'rgba(0, 0, 0, 0.85)';
    const childNodes = Array.from(rootRef.current.childNodes).filter(
      node => !!node.nodeType && node.classList.contains('microlight')
    );

    childNodes.forEach(node =>
      node.addEventListener('mousewheel', handlePreventYScrollingBeyondElement, { passive: false })
    );

    return () => {
      childNodes.forEach(node => node.removeEventListener('mousewheel', handlePreventYScrollingBeyondElement));
    };
  }, [value, className, language]);

  const handleDownload = () => {
    saveAs(value, fileName);
  };

  const handlePreventYScrollingBeyondElement = e => {
    const { target, deltaY } = e;
    const { scrollHeight: contentHeight, offsetHeight: visibleHeight, scrollTop } = target;
    const scrollOffset = visibleHeight + scrollTop;
    const isElementScrollable = contentHeight > visibleHeight;
    const isScrollingPastTop = scrollTop === 0 && deltaY < 0;
    const isScrollingPastBottom = scrollOffset >= contentHeight && deltaY > 0;

    if (isElementScrollable && (isScrollingPastTop || isScrollingPastBottom)) {
      e.preventDefault();
    }
  };

  return (
    <div className="highlight-code" ref={rootRef}>
      {!downloadable ? null : (
        <div className="download-contents" onClick={handleDownload}>
          <DownSvg style={{ width: '16px', height: '15px', minWidth: '16px' }} />
        </div>
      )}

      {canCopy && (
        <div className="copy-to-clipboard">
          <CopyToClipboard text={value}>
            <button onClick={() => message.success(intl.get('dataAgent.copyKeySuccess'))} />
          </CopyToClipboard>
        </div>
      )}

      {canSyntaxHighlight ? (
        <SyntaxHighlighter
          language={language}
          className={cx(className, 'microlight')}
          style={getStyle(get(config, 'syntaxHighlight.theme', 'agate'))}
        >
          {value}
        </SyntaxHighlighter>
      ) : (
        <pre
          className={cx(className, 'microlight')}
          style={{ background: 'rgba(0,0,0,0.04)', border: '1px solid rgba(0,0,0,0.1)', color: 'rgba(0,0,0,0.85)' }}
        >
          {value}
        </pre>
      )}
    </div>
  );
};

HighlightCode.propTypes = {
  value: PropTypes.string.isRequired,
  getConfigs: PropTypes.func.isRequired,
  className: PropTypes.string,
  downloadable: PropTypes.bool,
  fileName: PropTypes.string,
  language: PropTypes.string,
  canCopy: PropTypes.bool,
};

HighlightCode.defaultProps = {
  fileName: 'response.txt',
};

export default HighlightCode;
