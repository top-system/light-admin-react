/**
 * <AvatarUpload> — click-to-upload circular avatar. Posts to /files via the
 * shared upload service and hands the returned url back through `onChange`,
 * so it works as a controlled field inside ProForm / Form.Item.
 */
import { CameraOutlined, LoadingOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import { App, Avatar, Upload } from 'antd';
import React, { useState } from 'react';
import { uploadFile } from '@/services/light-admin/file';

export type AvatarUploadProps = {
  value?: string;
  onChange?: (url: string | undefined) => void;
  /** Fallback text shown when no avatar is set (usually the first char of the nickname). */
  fallback?: React.ReactNode;
  size?: number;
  maxSizeMb?: number;
  disabled?: boolean;
  /** Resolve a stored (possibly relative) url to something the browser can render. */
  resolveUrl?: (url: string) => string;
};

const ACCEPT = 'image/png,image/jpeg,image/gif,image/webp';

export const AvatarUpload: React.FC<AvatarUploadProps> = ({
  value,
  onChange,
  fallback,
  size = 96,
  maxSizeMb = 2,
  disabled,
  resolveUrl,
}) => {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState(false);

  const beforeUpload: UploadProps['beforeUpload'] = (file) => {
    if (!file.type.startsWith('image/')) {
      message.error('只能上传图片文件');
      return Upload.LIST_IGNORE;
    }
    if (file.size > maxSizeMb * 1024 * 1024) {
      message.error(`图片不能超过 ${maxSizeMb}MB`);
      return Upload.LIST_IGNORE;
    }
    return true;
  };

  const customRequest: UploadProps['customRequest'] = async ({
    file,
    onSuccess,
    onError,
  }) => {
    try {
      setUploading(true);
      const result = await uploadFile(file as File);
      onChange?.(result.url);
      onSuccess?.(result, file as never);
      message.success('头像已上传，保存后生效');
    } catch (err) {
      onError?.(err as Error);
      message.error('头像上传失败');
    } finally {
      setUploading(false);
    }
  };

  const src = value ? (resolveUrl ? resolveUrl(value) : value) : undefined;

  return (
    <Upload
      accept={ACCEPT}
      showUploadList={false}
      beforeUpload={beforeUpload}
      customRequest={customRequest}
      disabled={disabled || uploading}
    >
      <div
        className="avatar-upload"
        style={{
          position: 'relative',
          width: size,
          height: size,
          borderRadius: '50%',
          overflow: 'hidden',
          cursor: disabled ? 'not-allowed' : 'pointer',
        }}
      >
        <Avatar size={size} src={src} style={{ fontSize: size / 2.5 }}>
          {fallback}
        </Avatar>
        <div
          className="avatar-upload-mask"
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'rgba(0,0,0,0.45)',
            color: '#fff',
            fontSize: size / 4,
            opacity: uploading ? 1 : 0,
            transition: 'opacity .2s',
          }}
          onMouseEnter={(e) => {
            if (!disabled) e.currentTarget.style.opacity = '1';
          }}
          onMouseLeave={(e) => {
            if (!uploading) e.currentTarget.style.opacity = '0';
          }}
        >
          {uploading ? <LoadingOutlined /> : <CameraOutlined />}
        </div>
      </div>
    </Upload>
  );
};

export default AvatarUpload;
