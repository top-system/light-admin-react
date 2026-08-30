/**
 * Profile — the signed-in user's own page, reached from the header avatar
 * dropdown ("个人信息"). Left: avatar (click to upload) + account summary.
 * Right: tabs for editing basic info and changing the password.
 */
import {
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  UserOutlined,
} from '@ant-design/icons';
import {
  PageContainer,
  ProForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { App, Card, Col, Descriptions, Row, Tabs, Tag, Typography } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import AvatarUpload from '@/components/business/AvatarUpload';
import {
  changePassword,
  getProfile,
  updateProfile,
} from '@/services/light-admin/user';
import type {
  ChangePasswordRequest,
  CurrentUser,
  UserProfileUpdate,
} from '@/types/light-admin/domain';

type PasswordFormValues = ChangePasswordRequest & { confirmPassword: string };

const GENDER_OPTIONS = [
  { value: 0, label: '未知' },
  { value: 1, label: '男' },
  { value: 2, label: '女' },
];

const ProfilePage: React.FC = () => {
  const { message } = App.useApp();
  const { setInitialState } = useModel('@@initialState');
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [avatar, setAvatar] = useState<string | undefined>();
  const [savingAvatar, setSavingAvatar] = useState(false);
  const passwordFormRef = useRef<ProFormInstance<PasswordFormValues>>(undefined);

  const refresh = async () => {
    const fresh = await getProfile();
    setUser(fresh);
    setAvatar(fresh.avatar || undefined);
    setInitialState((s) => (s ? { ...s, currentUser: fresh } : s));
    return fresh;
  };

  useEffect(() => {
    getProfile().then((u) => {
      setUser(u);
      setAvatar(u.avatar || undefined);
    });
  }, []);

  // Avatar saves immediately on upload — no need to hit "保存" on the form.
  const handleAvatarChange = async (url: string | undefined) => {
    setAvatar(url);
    if (!url) return;
    try {
      setSavingAvatar(true);
      await updateProfile({ avatar: url });
      await refresh();
      message.success('头像已更新');
    } finally {
      setSavingAvatar(false);
    }
  };

  const handleSubmit = async (values: UserProfileUpdate) => {
    await updateProfile({
      ...values,
      gender: values.gender !== undefined ? Number(values.gender) : undefined,
      avatar,
    });
    await refresh();
    message.success('保存成功');
  };

  const handleChangePassword = async (values: PasswordFormValues) => {
    await changePassword({
      oldPassword: values.oldPassword,
      newPassword: values.newPassword,
    });
    message.success('密码修改成功，下次登录请使用新密码');
    passwordFormRef.current?.resetFields();
    return true;
  };

  if (!user) {
    return <PageContainer loading />;
  }

  return (
    <PageContainer title="个人信息">
      <Row gutter={[16, 16]}>
        <Col xs={24} md={8} lg={7} xl={6}>
          <Card>
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 8,
              }}
            >
              <AvatarUpload
                value={avatar}
                onChange={handleAvatarChange}
                fallback={user.nickname?.[0] || user.username?.[0]}
                size={104}
                disabled={savingAvatar}
              />
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                点击头像上传（≤2MB，png/jpg/gif/webp）
              </Typography.Text>
              <Typography.Title level={4} style={{ margin: '8px 0 0' }}>
                {user.nickname || user.username}
              </Typography.Title>
              <Typography.Text type="secondary">@{user.username}</Typography.Text>
            </div>
            <Descriptions
              column={1}
              size="small"
              style={{ marginTop: 24 }}
              items={[
                {
                  key: 'dept',
                  label: '部门',
                  children: user.deptName || '-',
                },
                {
                  key: 'roles',
                  label: '角色',
                  children: user.roles?.length
                    ? user.roles.map((r) => <Tag key={r}>{r}</Tag>)
                    : '-',
                },
                {
                  key: 'mobile',
                  label: '手机',
                  children: user.mobile || '-',
                },
                {
                  key: 'email',
                  label: '邮箱',
                  children: user.email || '-',
                },
                {
                  key: 'createTime',
                  label: '注册时间',
                  children: user.createTime || '-',
                },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} md={16} lg={17} xl={18}>
          <Card>
            <Tabs
              items={[
                {
                  key: 'basic',
                  label: '基本信息',
                  children: (
                    <ProForm<UserProfileUpdate>
                      initialValues={{
                        nickname: user.nickname,
                        mobile: user.mobile,
                        email: user.email,
                        gender: user.gender,
                      }}
                      onFinish={handleSubmit}
                      layout="vertical"
                      style={{ maxWidth: 480 }}
                      submitter={{
                        searchConfig: { submitText: '保存' },
                        resetButtonProps: { style: { display: 'none' } },
                      }}
                    >
                      <ProFormText
                        name="nickname"
                        label="昵称"
                        fieldProps={{ prefix: <UserOutlined /> }}
                        rules={[
                          { required: true, message: '请输入昵称' },
                          { max: 32, message: '昵称不能超过 32 个字符' },
                        ]}
                      />
                      <ProFormText
                        name="mobile"
                        label="手机号"
                        fieldProps={{ prefix: <PhoneOutlined /> }}
                        rules={[
                          {
                            pattern: /^1[3-9]\d{9}$/,
                            message: '请输入正确的手机号',
                          },
                        ]}
                      />
                      <ProFormText
                        name="email"
                        label="邮箱"
                        fieldProps={{ prefix: <MailOutlined /> }}
                        rules={[{ type: 'email', message: '请输入正确的邮箱' }]}
                      />
                      <ProFormSelect
                        name="gender"
                        label="性别"
                        options={GENDER_OPTIONS}
                      />
                    </ProForm>
                  ),
                },
                {
                  key: 'password',
                  label: '修改密码',
                  children: (
                    <ProForm<PasswordFormValues>
                      formRef={passwordFormRef}
                      onFinish={handleChangePassword}
                      layout="vertical"
                      style={{ maxWidth: 480 }}
                      submitter={{
                        searchConfig: { submitText: '修改密码' },
                        resetButtonProps: { style: { display: 'none' } },
                      }}
                    >
                      <ProFormText.Password
                        name="oldPassword"
                        label="当前密码"
                        fieldProps={{
                          prefix: <LockOutlined />,
                          autoComplete: 'current-password',
                        }}
                        rules={[{ required: true, message: '请输入当前密码' }]}
                      />
                      <ProFormText.Password
                        name="newPassword"
                        label="新密码"
                        fieldProps={{
                          prefix: <LockOutlined />,
                          autoComplete: 'new-password',
                        }}
                        rules={[
                          { required: true, message: '请输入新密码' },
                          { min: 6, message: '密码至少 6 位' },
                          { max: 64, message: '密码不能超过 64 位' },
                          ({ getFieldValue }) => ({
                            validator: (_, value) =>
                              value && value === getFieldValue('oldPassword')
                                ? Promise.reject(
                                    new Error('新密码不能与当前密码相同'),
                                  )
                                : Promise.resolve(),
                          }),
                        ]}
                      />
                      <ProFormText.Password
                        name="confirmPassword"
                        label="确认新密码"
                        dependencies={['newPassword']}
                        fieldProps={{
                          prefix: <LockOutlined />,
                          autoComplete: 'new-password',
                        }}
                        rules={[
                          { required: true, message: '请再次输入新密码' },
                          ({ getFieldValue }) => ({
                            validator: (_, value) =>
                              !value || value === getFieldValue('newPassword')
                                ? Promise.resolve()
                                : Promise.reject(
                                    new Error('两次输入的密码不一致'),
                                  ),
                          }),
                        ]}
                      />
                    </ProForm>
                  ),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default ProfilePage;
