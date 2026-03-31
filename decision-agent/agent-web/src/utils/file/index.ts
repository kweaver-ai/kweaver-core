export enum ImportFileErrorEnum {
  UserCancelled = 'userCancelled',
  ReadError = 'readError',
  SizeLimitExceeded = 'sizeLimitExceeded',
}

export enum ReadAsEnum {
  Text = 'text',
  DataURL = 'dataURL',
  ArrayBuffer = 'arrayBuffer',
}

interface ImportFileOptions {
  accept?: string;
  readAs?: ReadAsEnum;
  maxFileSize?: number;
}

interface ImportFileError {
  error: ImportFileErrorEnum;
  message?: string;
}

// 支持 JSON 对象和二进制数据（Blob/ArrayBuffer/Uint8Array）
export function downloadFile(
  data: unknown | Blob | ArrayBuffer | Uint8Array,
  filename: string,
  options?: { type?: string }
): void {
  let blob: Blob;
  if (data instanceof Blob) {
    blob = data; // 直接使用传入的 Blob
  } else if (data instanceof ArrayBuffer || data instanceof Uint8Array) {
    blob = new Blob([data], { type: options?.type || 'application/octet-stream' });
  } else {
    // 默认处理 JSON
    blob = new Blob([JSON.stringify(data, null, 2)], {
      type: options?.type || 'application/json',
    });
  }

  const link = document.createElement('a');
  const objectUrl = URL.createObjectURL(blob);
  link.href = objectUrl;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  // 立即释放 URL 对象，避免内存泄漏
  // 现代浏览器在点击下载后已经将数据复制到下载队列，可以安全释放
  URL.revokeObjectURL(objectUrl);
}

/**
 * 用于在网页中创建一个文件选择对话框，让用户选择一个文件（默认是JSON文件），然后读取并解析该文件内容。
 */
export async function importFile<T = any>({
  accept = '.json',
  readAs = ReadAsEnum.Text,
  maxFileSize,
}: ImportFileOptions = {}): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = accept;
    input.style.display = 'none';

    const cleanup = () => {
      document.body.removeChild(input);
      input.onchange = null;
    };

    input.onchange = async (event: Event) => {
      const file = (event.target as HTMLInputElement).files?.[0];

      if (!file) {
        cleanup();
        reject({ error: ImportFileErrorEnum.UserCancelled } as ImportFileError);
        return;
      }

      // 检查文件大小限制
      if (maxFileSize && file.size > maxFileSize) {
        cleanup();
        reject({ error: ImportFileErrorEnum.SizeLimitExceeded } as ImportFileError);
        return;
      }

      try {
        const reader = new FileReader();

        reader.onload = () => {
          cleanup();
          resolve(reader.result as T);
        };

        reader.onerror = () => {
          cleanup();
          reject({
            error: ImportFileErrorEnum.ReadError,
            message: reader.error?.message || 'Failed to read file',
          } as ImportFileError);
        };

        // 根据不同的读取方式调用不同的方法
        switch (readAs) {
          case ReadAsEnum.DataURL:
            reader.readAsDataURL(file);
            break;
          case ReadAsEnum.ArrayBuffer:
            reader.readAsArrayBuffer(file);
            break;
          default:
            reader.readAsText(file);
        }
      } catch (error) {
        cleanup();
        reject({
          error: ImportFileErrorEnum.ReadError,
          message: error instanceof Error ? error.message : String(error),
        } as ImportFileError);
      }
    };

    document.body.appendChild(input);
    input.click();
  });
}

// 解析文件名
export function getFilenameFromContentDisposition(contentDisposition: string) {
  if (!contentDisposition) return '';

  const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);

  if (filenameMatch && filenameMatch[1]) {
    // 去除可能存在的引号
    let filename = filenameMatch[1].replace(/['"]/g, '');
    // 处理可能的编码字符（如UTF-8编码的文件名）
    try {
      filename = decodeURIComponent(filename);
    } catch {
      // 如果解码失败，保持原样
    }
    return filename;
  }

  return '';
}
