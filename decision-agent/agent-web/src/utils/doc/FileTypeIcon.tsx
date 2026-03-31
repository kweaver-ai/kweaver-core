import { useMemo } from 'react';
import { getFileIcon } from './file-icon';
import { getFileExtension } from './extension';

interface PropsType {
  extension?: string;
  className?: string;
  fontSize?: number | string;
  size?: number;
  style?: Record<string, any>;
  name?: string;
}

function FileTypeIcon({ extension, className, fontSize = 16, size, style, name }: PropsType) {
  const Icon = useMemo(() => {
    const ext = extension || getFileExtension(name || '');
    return getFileIcon({ size, extension: ext });
  }, [size, extension, name]);

  const width = typeof fontSize === 'string' ? fontSize : `${fontSize}px`;

  return (
    <div className={className} style={{ fontSize, ...style }}>
      <Icon style={{ fontSize, width, height: width }} />
    </div>
  );
}

export default FileTypeIcon;
