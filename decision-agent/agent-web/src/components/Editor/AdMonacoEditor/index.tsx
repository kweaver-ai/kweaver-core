import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';
import type { DiffEditorProps, EditorProps, Monaco } from '@monaco-editor/react';
import Editor, { DiffEditor } from '@monaco-editor/react';
import classNames from 'classnames';
import './styles.less';
// import { initAdMonacoEditor } from './assitants';
import { Chrome_DevTools_Theme } from '../static';
import useLatestState from '@/hooks/useLatestState';

// initAdMonacoEditor();

type AdMonacoEditorCommonProps = {
  bordered?: boolean;
  placeholder?: string;
  className?: string;
  style?: React.CSSProperties;
  type?: 'editor' | 'diff-editor';
  autoFocus?: boolean;
  onFocus?: () => void;
  onBlur?: () => void;
  onDidContentSizeChange?: (editor: any) => void;
  onMouseUp?: (event: any) => void;
  onMouseWheel?: (event: any) => void;
  onError?: (error: any[]) => void;
};

type IEditorProps = EditorProps & AdMonacoEditorCommonProps;
type IDiffEditorProps = DiffEditorProps & AdMonacoEditorCommonProps;

export interface AdMonacoEditorProps extends IEditorProps {
  minHeight?: number; // 配合 height: 'auto' 使用， 用于设置自动高度时的最小高度
  maxHeight?: number; // 配合 height: 'auto' 使用， 用于设置自动高度时的最大高度
  autoScrollBottom?: boolean; //  是否自动滚动到代码底部
  onHeightChange?: (height: number) => void; // 高度值变化事件
}

export interface AdMonacoDiffEditorProps extends IDiffEditorProps {
  originalVisible?: boolean;
}

export type AdMonacoEditorRef = {
  getMonacoInstance: () => Monaco;
  getEditorInstance: () => any;
};

type MonacoEditorProps = AdMonacoEditorProps & AdMonacoDiffEditorProps;
const MonacoEditor = forwardRef<AdMonacoEditorRef, MonacoEditorProps>((props, ref) => {
  const {
    className,
    placeholder = '请输入',
    placeholderStyle,
    style,
    theme,
    type = 'editor',
    height = '100%',
    minHeight,
    maxHeight,
    autoFocus = false,
    originalVisible = true,
    onMount,
    onChange,
    options,
    bordered = false,
    autoScrollBottom = false,
    onFocus,
    onBlur,
    onMouseUp,
    onMouseWheel,
    onError,
    onDidContentSizeChange,
    onHeightChange,
    ...restMonacoProps
  } = props;
  const [loaded, setLoaded] = useState<boolean>(false);
  const [editorHeight, setEditorHeight, getEditorHeight] = useLatestState(
    height === 'auto' ? (minHeight ? minHeight : 22) : height
  ); // 初始高度
  const monacoRef = useRef<Monaco>();
  const editorRef = useRef<any>();
  const maxHeightRef = useRef(maxHeight);
  const minHeightRef = useRef(minHeight);
  const canScrollBottom = useRef(true);

  maxHeightRef.current = maxHeight;
  minHeightRef.current = minHeight;

  useImperativeHandle(ref, () => ({
    getMonacoInstance: () => monacoRef.current,
    getEditorInstance: () => editorRef.current,
  }));

  // 设置主题
  useEffect(() => {
    const monaco = monacoRef.current;
    if (theme && monaco && loaded) {
      monaco.editor.setTheme(theme);
    }
  }, [theme, loaded]);

  useEffect(() => {
    if (loaded) {
      if (height === 'auto') {
        const contentHeight = editorRef.current.getContentHeight();
        updateEditorHeight(contentHeight);
      } else {
        setEditorHeight(`${height}px`);
        editorRef.current.layout();
      }
    }
  }, [height, loaded, minHeight, maxHeight]);

  const updateEditorHeight = (contentHeight: number) => {
    if (editorRef.current) {
      const scrollBarHeight = 16; // 水平方向滚动条高度
      let height = contentHeight + scrollBarHeight;
      if (minHeightRef.current && contentHeight < minHeightRef.current) {
        height = minHeightRef.current;
      }
      if (maxHeightRef.current && contentHeight > maxHeightRef.current) {
        height = maxHeightRef.current;
      }
      onHeightChange?.(height);
      setEditorHeight(`${height}px`);
      editorRef.current.layout();
    }
  };

  /** 自动滚动到最底部 */
  const handleAutoScrollBottom = (editor: any) => {
    if (canScrollBottom.current) {
      const lineCount = editor.getModel()?.getLineCount() || 0;
      editor.revealLineInCenterIfOutsideViewport(lineCount, 0);
    }
  };

  const handleMount: EditorProps['onMount'] = (editor, monaco) => {
    if (autoFocus) {
      editor.focus();
    }
    monacoRef.current = monaco;
    editorRef.current = editor;
    setLoaded(true);
    onMount?.(editor, monaco);

    editor.onDidFocusEditorWidget(() => {
      onFocus?.();
    });
    editor.onDidBlurEditorWidget(() => {
      onBlur?.();
    });
    editor.onMouseUp((event: any) => {
      onMouseUp?.(event);
    });
    editor.onMouseWheel((event: any) => {
      canScrollBottom.current = false;
      onMouseWheel?.(event);
    });
    editor.onDidContentSizeChange((event: any) => {
      if (height === 'auto') {
        updateEditorHeight(event.contentHeight);
        if (autoScrollBottom && maxHeightRef.current && event.contentHeight > maxHeightRef.current) {
          handleAutoScrollBottom(editor);
        }
      }
      onDidContentSizeChange?.(editor);
    });
    editor.onKeyDown((e: any) => {
      if (e.keyCode === monaco.KeyCode.Enter) {
        // e.stopPropagation(); // 阻止事件冒泡(此处这样做的目的是解决编辑器放在antd的FormItem里面之后回车键不能换行的问题
      }
    });

    monaco.editor.onDidChangeMarkers(([resource]: any) => {
      const model = editor.getModel();
      if (model && resource.toString() === model.uri.toString()) {
        const markers = monaco.editor.getModelMarkers({ resource });
        const errors = markers.filter((marker: any) => marker.severity === monaco.MarkerSeverity.Error);
        if (errors.length > 0) {
          onError?.(
            errors.map((error: any) => ({
              message: error.message,
              line: error.startLineNumber,
              column: error.startColumn,
            }))
          );
        } else {
          onError?.([]);
        }
      }
    });
  };

  const handelChange: EditorProps['onChange'] = (value, ev) => {
    onChange?.(value, ev);
  };

  const mergeOptions = {
    ...options,
    lineHeight: options.lineHeight ?? 22,
    scrollBeyondLastLine: height === 'auto' ? false : options?.scrollBeyondLastLine,
  };

  return (
    <div
      className={classNames('ad-monaco-editor dip-w-100 dip-h-100', className, {
        'ad-monaco-diff-editor': type === 'diff-editor',
        'ad-monaco-diff-editor-original-hidden': !originalVisible,
        'ad-monaco-editor-border': bordered,
      })}
      style={style}
    >
      {!restMonacoProps.value && placeholder && (
        <div className="ad-monaco-editor-placeholder" style={placeholderStyle}>
          {placeholder}
        </div>
      )}
      {type === 'editor' ? (
        <Editor
          height={editorHeight}
          onMount={handleMount}
          onChange={handelChange}
          options={mergeOptions}
          {...restMonacoProps}
        />
      ) : (
        <DiffEditor height={editorHeight} onMount={handleMount} options={mergeOptions} {...restMonacoProps} />
      )}
    </div>
  );
});

MonacoEditor.defaultProps = {
  height: '100%',
  defaultLanguage: 'json',
  theme: Chrome_DevTools_Theme,
};

export const AdMonacoEditor = forwardRef<AdMonacoEditorRef, AdMonacoEditorProps>((props, ref) => (
  <MonacoEditor type="editor" {...props} ref={ref} />
));

export const AdMonacoDiffEditor = forwardRef<AdMonacoEditorRef, AdMonacoDiffEditorProps>((props, ref) => (
  <MonacoEditor type="diff-editor" {...props} ref={ref} />
));
