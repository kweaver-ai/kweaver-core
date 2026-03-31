import React, { useMemo, useRef, useState } from 'react';
import { uniqBy } from 'lodash';
import './style.less';
import { AdMonacoEditor } from '../AdMonacoEditor';
import { Dolphin_Theme } from '../static';
import useDeepCompareEffect from '@/hooks/useDeepCompareEffect';
import {
  extractCustomVarFromBeforeCursor,
  // extractToolsFromImports,
  registerBlockVarCompletion,
  registerCallToolCompletion,
  // registerImportToolCompletion,
  registerKeywordsCompletion,
  triggerSuggestByChar,
} from './utils';
import classNames from 'classnames';

export type AdDolphinEditorProps = {
  className?: string;
  style?: React.CSSProperties;
  value?: string;
  placeholder?: string;
  placeholderStyle?: Record<string, any>;
  promptVarOptions: Array<{ label: string; value: string }>;
  onChange?: (value?: string) => void;
  onFocus?: () => void;
  onBlur?: () => void;
  onError?: (error: any[]) => void;
  onMouseUp?: (e: any) => void;
  onMount?: (editor: any, monaco: any) => void;
  disabled?: boolean;
  showLineNumbers?: boolean;
  toolOptions?: string[];
  height?: number | 'auto';
  minHeight?: number | undefined;
  maxHeight?: number | undefined;
  onHeightChange?: (height: number) => void; // 高度值变化事件
};

const AdDolphinEditor = (props: AdDolphinEditorProps) => {
  const {
    promptVarOptions = [],
    toolOptions = [],
    placeholder = '请输入 dolphin language',
    placeholderStyle,
    value,
    disabled = false,
    showLineNumbers = true,
    height = 'auto',
    minHeight,
    maxHeight,
    onChange,
    onFocus,
    onBlur,
    onMount,
    onMouseUp,
    onError,
    onHeightChange,
  } = props;
  const [rendered, setRendered] = useState(false);
  const [focused, setFocused] = useState(false);
  const editorRef = useRef<any>(null);
  const monacoRef = useRef<any>(null);
  const prefixCls = 'AdDolphinEditor';
  // const tools = useMemo(() => {
  //   return extractToolsFromImports(value || '');
  // }, [value]);

  const customVar = useMemo(() => {
    if (editorRef.current) {
      return extractCustomVarFromBeforeCursor(value || '', editorRef.current!);
    }
    return [];
  }, [value]);

  // 聚焦的时候注册补全项， 失焦的时候销毁补全项，防止补权项重复
  useDeepCompareEffect(() => {
    // let importToolCompletionItemProvider: any;
    let keywordsCompletionItemProvider: any;
    const monaco = monacoRef.current;
    if (monaco && rendered && focused) {
      // 注册import工具动态智能提示
      // importToolCompletionItemProvider = registerImportToolCompletion(monaco);
      keywordsCompletionItemProvider = registerKeywordsCompletion(monaco);
    }
    return () => {
      // 清除上一次注册的智能提示，防止出现重复的提示（重复注册会出现重复的提示）
      // importToolCompletionItemProvider?.dispose();
      // importToolCompletionItemProvider = null;

      keywordsCompletionItemProvider?.dispose();
      keywordsCompletionItemProvider = null;
    };
  }, [rendered, focused]);

  useDeepCompareEffect(() => {
    let blockVarCompletionItemProvider: any;
    const monaco = monacoRef.current;
    if (monaco && rendered && focused) {
      // 注册变量动态智能提示
      if (promptVarOptions.length > 0) {
        blockVarCompletionItemProvider = registerBlockVarCompletion(
          monaco,
          uniqBy([...promptVarOptions, ...customVar], 'value')
        );
      }
    }
    return () => {
      // 清除上一次注册的智能提示，防止出现重复的提示（重复注册会出现重复的提示）
      blockVarCompletionItemProvider?.dispose();
      blockVarCompletionItemProvider = null;
    };
  }, [rendered, focused, customVar, promptVarOptions]);

  useDeepCompareEffect(() => {
    let callToolCompletionItemProvider: any;
    const monaco = monacoRef.current;
    if (monaco && rendered && focused) {
      // 从value中提取import进来的工具
      callToolCompletionItemProvider = registerCallToolCompletion(monaco, toolOptions);
    }
    return () => {
      callToolCompletionItemProvider?.dispose();
      callToolCompletionItemProvider = null;
    };
    // }, [rendered, focused, tools]);
  }, [rendered, focused, toolOptions]);

  return (
    <AdMonacoEditor
      placeholder={placeholder}
      placeholderStyle={placeholderStyle}
      className={classNames(prefixCls, {
        [`${prefixCls}-disabled`]: disabled,
      })}
      value={value}
      height={height}
      minHeight={minHeight}
      maxHeight={maxHeight}
      options={{
        minimap: { enabled: false },
        fontSize: 14,
        lineHeight: 22,
        tabSize: 2,
        insertSpaces: true,
        readOnly: disabled,
        lineNumbers: showLineNumbers ? 'on' : 'off',
        scrollbar: {
          alwaysConsumeMouseWheel: false, // 禁用Monaco的默认滚轮事件
        },
        lineNumbersMinChars: 4,
        unicodeHighlight: {
          ambiguousCharacters: false, // 关闭中文符号高亮报警
        },
        suggest: {
          showWords: false,
        },
        scrollBeyondLastLine: false,
      }}
      defaultLanguage="dolphin"
      theme={Dolphin_Theme}
      beforeMount={() => {
        editorRef.current?.getMonacoInstance()?.dispose();
      }}
      onMount={(editor, monaco) => {
        editorRef.current = editor;
        monacoRef.current = monaco;
        setRendered(true);
        onMount?.(editor, monaco);
      }}
      onChange={value => {
        onChange?.(value);
        if (editorRef.current) {
          triggerSuggestByChar(editorRef.current);
        }
      }}
      onFocus={() => {
        onFocus?.();
        setFocused(true);
      }}
      onBlur={() => {
        onBlur?.();
        setFocused(false);
      }}
      onMouseUp={(e: any) => {
        onMouseUp?.(e);
      }}
      onError={onError}
      onHeightChange={onHeightChange}
    />
  );
};

export default AdDolphinEditor;
