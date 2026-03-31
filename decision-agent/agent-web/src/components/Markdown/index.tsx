import React from 'react';
import { MdEditor, MdPreview } from 'md-editor-rt';
import 'md-editor-rt/lib/style.css';
// import 'md-editor-rt/lib/preview.css';
import classNames from 'classnames';
import './index.less';
import { initMarkdownConfig } from './utils';

type MarkdownProps = {
  value?: string;
  onChange?: (value: string) => void;
  readOnly?: boolean;
  className?: string;
};

initMarkdownConfig();
const Markdown: React.FC<MarkdownProps> = props => {
  const { className, value, onChange, readOnly = false } = props;
  const prefixCls = 'dip-markdown';
  const renderContent = () => {
    if (readOnly) {
      return (
        <>
          <MdPreview noEcharts previewTheme="github" autoFoldThreshold={1000000000} value={value} />
        </>
      );
    }
    return (
      <MdEditor
        codeFoldable={false}
        previewTheme="github"
        toolbars={[
          'bold',
          'underline',
          'italic',
          '-',
          'strikeThrough',
          'sub',
          'sup',
          'quote',
          'unorderedList',
          'orderedList',
          'task',
          '-',
          'codeRow',
          'code',
          'link',
          // 'image',
          'table',
          'mermaid',
          'katex',
          // '-',
          // 'revoke',
          // 'next',
          // 'save',
          '=',
          0,
          // 'pageFullscreen',
          // 'fullscreen',
          // 'preview',
          // 'previewOnly',
          // 'htmlPreview',
          'catalog',
          // 'github',
        ]}
        language="zh-CN"
        value={value}
        onChange={onChange}
      />
    );
  };

  return <div className={classNames(`${prefixCls}`, className)}>{renderContent()}</div>;
};

export default Markdown;
