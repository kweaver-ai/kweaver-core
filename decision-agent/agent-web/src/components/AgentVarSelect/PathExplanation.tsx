import { memo, FC } from 'react';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { InfoCircleOutlined } from '@ant-design/icons';
import styles from './PathExplanation.module.css';

interface PathExplanationProps {
  isRoot?: boolean;
}

const PathExplanation: FC<PathExplanationProps> = ({ isRoot }) => {
  return (
    <div
      className={classNames(
        'dip-mt-8 dip-border-radius-4 dip-font-12',
        styles['path-explanation'],
        styles['grey-color1']
      )}
    >
      <div className="dip-user-select-none">
        <InfoCircleOutlined className="dip-mr-6 dip-font-14" />
        {intl.get('dataAgent.pathDescription')}
      </div>
      {isRoot ? (
        <div className="dip-user-select-none">{intl.get('dataAgent.selectRootPathTip')}</div>
      ) : (
        <>
          <div className="dip-user-select-none">{intl.get('dataAgent.selectArraySubfieldTip')}</div>
          <div className={classNames('dip-mt-8 dip-border-radius-4 dip-font-11', styles['demo'])}>
            <div className={classNames('dip-mb-4 dip-c-bold', styles['grey-color2'])}>
              {intl.get('dataAgent.example')}
            </div>
            <div className={classNames('dip-mb-4', styles['grey-color1'])}>{intl.get('dataAgent.originalData')}</div>
            <pre className={classNames('dip-mb-6 dip-font-10', styles['code-1'], styles['grey-color2'])}>{`{
  "llms": [
    { "is_default": true },
    { "is_default": false }
  ]
}`}</pre>
            <div className={classNames('dip-font-10', styles['grey-color1'])}>
              <code className={classNames('dip-border-radius-2', styles['code-2'])}>
                self_config.llms[*].is_default
              </code>
              <span className={classNames(styles['code-3'], styles['grey-color1'])}>→</span>
              <code className={classNames('dip-border-radius-2', styles['code-4'])}>[true, false]</code>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default memo(PathExplanation);
