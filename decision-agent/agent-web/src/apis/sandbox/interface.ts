export type SandBoxUploadFileProps = {
  file: File;
  sessionId: string;
  filePath: string;
}

export type SandBoxQueryFileProps = {
  sessionId: string;
  path: string;
  limit: number;
}