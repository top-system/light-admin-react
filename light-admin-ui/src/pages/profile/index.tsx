/**
 * Profile — tabs: basic info + avatar upload. Password change endpoint isn't
 * part of the current contract so we skip that tab.
 */
import {
  PageContainer,
  ProForm,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { App, Avatar, Card, Tabs } from 'antd';
import React, { useEffect, useState } from 'react';
import FileUpload from '@/components/business/FileUpload';
import { getProfile, updateProfile } from '@/services/light-admin/user';
import type {
  CurrentUser,
  UserProfileUpdate,
} from '@/types/light-admin/domain';

const ProfilePage: React.FC = () => {
  const { message } = App.useApp();
  const { setInitialState } = useModel('@@initialState');
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [avatar, setAvatar] = useState<string | undefined>();

  useEffect(() => {
    getProfile().then((u) => {
      setUser(u);
      setAvatar(u.avatar);
    });
  }, []);

  const handleSubmit = async (values: UserProfileUpdate) => {
    await updateProfile({ ...values, avatar });
    message.success('保存成功');
    const fresh = await getProfile();
    setUser(fresh);
    setInitialState((s) => (s ? { ...s, currentUser: fresh } : s));
  };

  return (
    <PageContainer>
      <Card>
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: user ? (
                <ProForm<UserProfileUpdate>
                  initialValues={user}
                  onFinish={handleSubmit}
                  layout="vertical"
                >
                  <Avatar size={64} src={avatar} style={{ marginBottom: 12 }}>
                    {user.nickname?.[0]}
                  </Avatar>
                  <ProForm.Item label="头像">
                    <FileUpload
                      value={avatar}
                      onChange={setAvatar}
                      accept="image/*"
                      maxSizeMb={5}
                      buttonText="上传头像"
                    />
                  </ProForm.Item>
                  <ProFormText
                    name="nickname"
                    label="昵称"
                    rules={[{ required: true }]}
                  />
                  <ProFormText name="mobile" label="手机号" />
                  <ProFormText name="email" label="邮箱" />
                  <ProFormSelect
                    name="gender"
                    label="性别"
                    options={[
                      { value: 0, label: '未知' },
                      { value: 1, label: '男' },
                      { value: 2, label: '女' },
                    ]}
                  />
                </ProForm>
              ) : null,
            },
          ]}
        />
      </Card>
    </PageContainer>
  );
};

export default ProfilePage;
