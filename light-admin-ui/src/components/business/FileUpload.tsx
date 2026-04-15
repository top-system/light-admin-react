/**
 * <FileUpload> — thin wrapper over antd Upload that posts to /files and
 * reports progress. Controlled by the `value` prop (a url string) so it
 * plays nicely inside ProForm / Form.Item.
 */
import { UploadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import { Button, message, Upload } from 'antd';
import React, { useState } from 'react';
import { uploadFile } from '@/services/light-admin/file';

export type FileUploadProps = {
  value?: string;
  onChange?: (url: string | undefined) => void;
  accept?: string;
  maxSizeMb?: number;
  buttonText?: string;
};

export const FileUpload: React.FC<FileUploadProps> = ({
  value,
  onChange,
  accept,
  maxSizeMb = 10,
  buttonText = '上传',
}) => {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const customRequest: UploadProps['customRequest'] = async ({
    file,
    onSuccess,
    onError,
  }) => {
    if (maxSizeMb && (file as File).size > maxSizeMb * 1024 * 1024) {
      message.error(`文件超过 ${maxSizeMb}MB`);
      onError?.(new Error('size limit'));
      return;
    }
    try {
      setUploading(true);
      setProgress(0);
      const result = await uploadFile(file as File, {
        onProgress: (p) => setProgress(p),
      });
      onChange?.(result.url);
      onSuccess?.(result, file as never);
    } catch (err) {
      onError?.(err as Error);
      message.error('上传失败');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <Upload
        accept={accept}
        showUploadList={false}
        customRequest={customRequest}
      >
        <Button icon={<UploadOutlined />} loading={uploading}>
          {uploading ? `${progress}%` : value ? '替换' : buttonText}
        </Button>
      </Upload>
      {value && (
        <div style={{ marginTop: 8 }}>
          <a href={value} target="_blank" rel="noreferrer">
            {value}
          </a>
          <Button
            type="link"
            size="small"
            onClick={() => onChange?.(undefined)}
          >
            移除
          </Button>
        </div>
      )}
    </div>
  );
};

export default FileUpload;
