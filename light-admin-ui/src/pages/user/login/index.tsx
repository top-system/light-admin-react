/**
 * Light Admin login page.
 *
 * Form layout: username, password, image captcha with click-to-refresh. On
 * success we persist the token, re-run getInitialState (so the menu/identity
 * are available before navigation), and push to `?redirect=...` or `/`.
 */
import { LockOutlined, SafetyOutlined, UserOutlined } from '@ant-design/icons';
import { LoginForm, ProFormText } from '@ant-design/pro-components';
import { Helmet } from '@umijs/max';
import { App, Image } from 'antd';
import { createStyles } from 'antd-style';
import React, { useCallback, useEffect, useState } from 'react';
import { Footer } from '@/components';
import { getCaptcha, login } from '@/services/light-admin/auth';
import type { Captcha } from '@/types/light-admin/domain';
import { safeRedirectTarget } from '@/utils/auth/redirect';
import { setToken } from '@/utils/auth/token';
import { ApiError } from '@/utils/response/adapter';

const useStyles = createStyles(({ token }) => ({
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100vh',
    overflow: 'auto',
    backgroundImage:
      "url('https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/V-_oS6r-i7wAAAAAAAAAAAAAFl94AQBr')",
    backgroundSize: '100% 100%',
  },
  captchaWrap: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
  },
  captchaImage: {
    cursor: 'pointer',
    height: 40,
    width: 120,
    borderRadius: token.borderRadius,
    border: `1px solid ${token.colorBorder}`,
    objectFit: 'cover',
  },
}));

type FormValues = {
  username: string;
  password: string;
  captchaCode: string;
};

const Login: React.FC = () => {
  const { styles } = useStyles();
  const { message } = App.useApp();
  const [captcha, setCaptcha] = useState<Captcha | null>(null);
  const [captchaLoading, setCaptchaLoading] = useState(false);

  const refreshCaptcha = useCallback(async () => {
    setCaptchaLoading(true);
    try {
      const next = await getCaptcha();
      setCaptcha(next);
    } catch (err) {
      const m = err instanceof ApiError ? err.message : '验证码加载失败';
      message.error(m);
    } finally {
      setCaptchaLoading(false);
    }
  }, [message]);

  useEffect(() => {
    void refreshCaptcha();
  }, [refreshCaptcha]);

  const handleSubmit = async (values: FormValues) => {
    if (!captcha) {
      message.error('请等待验证码加载');
      return;
    }
    try {
      const resp = await login({
        username: values.username,
        password: values.password,
        captchaId: captcha.captchaId,
        captchaCode: values.captchaCode,
      });
      setToken(resp.accessToken);
      message.success('登录成功');

      // Full-page navigation so getInitialState re-runs with the fresh token.
      // history.replace triggers onPageChange before the new state propagates,
      // and the layout's auth check bounces us right back to login.
      const redirect = new URL(window.location.href).searchParams.get(
        'redirect',
      );
      window.location.href = safeRedirectTarget(redirect);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : '登录失败，请重试';
      message.error(msg);
      void refreshCaptcha();
    }
  };

  return (
    <div className={styles.container}>
      <Helmet>
        <title>登录 - Light Admin</title>
      </Helmet>
      <div style={{ flex: 1, padding: '32px 0' }}>
        <LoginForm<FormValues>
          contentStyle={{ minWidth: 280, maxWidth: '75vw' }}
          logo={<img alt="logo" src="/logo.svg" />}
          title="Light Admin"
          subTitle="后台管理"
          onFinish={async (values) => {
            await handleSubmit(values);
          }}
          submitter={{
            searchConfig: { submitText: '登录' },
          }}
        >
          <ProFormText
            name="username"
            fieldProps={{ size: 'large', prefix: <UserOutlined /> }}
            placeholder="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          />
          <ProFormText.Password
            name="password"
            fieldProps={{ size: 'large', prefix: <LockOutlined /> }}
            placeholder="密码"
            rules={[{ required: true, message: '请输入密码' }]}
          />
          <div className={styles.captchaWrap}>
            <ProFormText
              name="captchaCode"
              fieldProps={{ size: 'large', prefix: <SafetyOutlined /> }}
              placeholder="验证码"
              formItemProps={{ style: { flex: 1, margin: 0 } }}
              rules={[{ required: true, message: '请输入验证码' }]}
            />
            <button
              type="button"
              onClick={() => void refreshCaptcha()}
              style={{ border: 'none', background: 'transparent', padding: 0 }}
              title="点击刷新验证码"
            >
              {captcha ? (
                <Image
                  src={captcha.captchaBase64}
                  alt="captcha"
                  preview={false}
                  className={styles.captchaImage}
                />
              ) : (
                <div
                  className={styles.captchaImage}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {captchaLoading ? '加载中…' : '点击获取'}
                </div>
              )}
            </button>
          </div>
        </LoginForm>
      </div>
      <Footer />
    </div>
  );
};

export default Login;
