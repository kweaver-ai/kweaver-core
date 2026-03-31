import intl from 'react-intl-universal';
import { AgentDetailType } from '@/apis/agent-factory/type';
import AdDolphinEditor from '@/components/Editor/AdDolphinEditor';
import { useMemo } from 'react';
import classNames from 'classnames';
import styles from './DolphinViewer.module.less';

interface DolphinDetailProps {
  config: AgentDetailType | null;
}

const DolphinViewer = ({ config }: DolphinDetailProps) => {
  const value = useMemo(() => {
    // 用户输入的dolphin value
    let userDolphinValue = '';
    // pre_dolphin的value
    let preDolphinValue = '';
    // post_dolphin的value
    let postDolphinValue = '';

    if (config?.config?.pre_dolphin) {
      config.config.pre_dolphin
        .filter(item => item.enabled)
        .forEach(item => {
          // 在模块上方添加注释：# <name>；如果模块value开头没有换行，需要手动添加换行
          preDolphinValue = preDolphinValue + `# ${item.name}${item.value.startsWith('\n') ? '' : '\n'}${item.value}\n`;
        });

      if (preDolphinValue) {
        preDolphinValue = `# ${intl.get('dataAgent.systemBuiltIn')}\n` + preDolphinValue;
      }
    }

    if (config?.config?.post_dolphin) {
      config.config.post_dolphin
        .filter(item => item.enabled)
        .forEach(item => {
          postDolphinValue =
            postDolphinValue + `# ${item.name}${item.value.startsWith('\n') ? '' : '\n'}${item.value}\n`;
        });

      if (postDolphinValue) {
        postDolphinValue = `# ${intl.get('dataAgent.systemBuiltIn')}\n` + postDolphinValue;
      }
    }

    if (config?.config?.dolphin) {
      if (preDolphinValue || postDolphinValue) {
        // 当开启了模块时，用户输入的dolphin上方添加注释 用于区分
        userDolphinValue = `# ${intl.get('dataAgent.userInput')}\n${config?.config?.dolphin}\n`;
      } else {
        userDolphinValue = config?.config?.dolphin;
      }
    }

    // 去除末尾的多个空行
    return (preDolphinValue + userDolphinValue + postDolphinValue).replace(/\n+$/, '');
  }, [config]);

  return (
    <div className={classNames(styles['dolphin-viewer'], 'dip-border-radius-8')}>
      <AdDolphinEditor
        value={value}
        placeholder={intl.get('dataAgent.noDescriptionAvailable')}
        disabled={true}
        promptVarOptions={[]}
        toolOptions={[]}
        maxHeight={394}
      />
    </div>
  );
};

export default DolphinViewer;
