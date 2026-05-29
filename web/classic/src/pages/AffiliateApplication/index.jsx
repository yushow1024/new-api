/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useState, useContext } from 'react';
import {
  Typography,
  Card,
  Button,
  Form,
  Input,
  TextArea,
  Row,
  Col,
  Toast,
  Space,
  Spin,
} from '@douyinfe/semi-ui';
import { IconMail } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { StatusContext } from '../../context/Status';

const { Title, Text, Paragraph } = Typography;

const AffiliateApplication = () => {
  const { t } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [loading, setLoading] = useState(false);
  const [formValues, setFormValues] = useState({
    name: '',
    email: '',
    phone: '',
    domain: '',
    expected_volume: '',
    business_description: '',
  });

  const handleInputChange = (field, value) => {
    setFormValues((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async () => {
    if (!formValues.name || !formValues.email) {
      showError(t('请填写必填项'));
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formValues.email)) {
      showError(t('请输入有效的邮箱地址'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/affiliate/application', formValues);
      if (res.data.success) {
        showSuccess(t('申请提交成功，我们将在 24 小时内与您联系'));
        setFormValues({
          name: '',
          email: '',
          phone: '',
          domain: '',
          expected_volume: '',
          business_description: '',
        });
      } else {
        showError(res.data.message || t('提交失败，请重试'));
      }
    } catch (error) {
      showError(error?.response?.data?.message || t('提交失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className='w-full overflow-x-hidden' style={{ marginTop: '64px' }}>
      <div
        style={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: 'white',
          padding: '60px 20px',
          textAlign: 'center',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background:
              'radial-gradient(circle at 80% 50%, rgba(255,255,255,0.1) 0%, transparent 50%)',
            pointerEvents: 'none',
          }}
        />
        <div style={{ position: 'relative', zIndex: 1 }}>
          <Title
            heading={1}
            style={{
              color: 'white',
              fontSize: 'clamp(1.8rem, 4vw, 2.5rem)',
              fontWeight: 700,
              marginBottom: '12px',
            }}
          >
            {t('代理申请')}
          </Title>
          <Text style={{ color: 'rgba(255,255,255,0.9)', fontSize: '1.1rem' }}>
            {t('填写以下信息，我们将在 24 小时内与您联系')}
          </Text>
        </div>
      </div>

      <div style={{ maxWidth: '800px', margin: '0 auto', padding: '40px 20px' }}>
        <Card
          style={{
            borderRadius: '16px',
            border: '1px solid var(--semi-color-border)',
            boxShadow: '0 4px 24px rgba(0,0,0,0.06)',
          }}
          bodyStyle={{ padding: '40px' }}
        >
          <Form layout='vertical'>
            <Row gutter={[24, 24]}>
              <Col span={12}>
                <Form.Label required>{t('姓名')}</Form.Label>
                <Input
                  placeholder={t('请输入您的姓名')}
                  value={formValues.name}
                  onChange={(value) => handleInputChange('name', value)}
                  style={{ width: '100%' }}
                />
              </Col>
              <Col span={12}>
                <Form.Label required>{t('邮箱')}</Form.Label>
                <Input
                  placeholder={t('请输入您的邮箱')}
                  value={formValues.email}
                  onChange={(value) => handleInputChange('email', value)}
                  style={{ width: '100%' }}
                  prefix={<IconMail />}
                />
              </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: '24px' }}>
              <Col span={12}>
                <Form.Label>{t('电话')}</Form.Label>
                <Input
                  placeholder={t('请输入您的联系电话')}
                  value={formValues.phone}
                  onChange={(value) => handleInputChange('phone', value)}
                  style={{ width: '100%' }}
                />
              </Col>
              <Col span={12}>
                <Form.Label>{t('域名')}</Form.Label>
                <Input
                  placeholder={t('例如：api.example.com')}
                  value={formValues.domain}
                  onChange={(value) => handleInputChange('domain', value)}
                  style={{ width: '100%' }}
                />
              </Col>
            </Row>

            <div style={{ marginTop: '24px' }}>
              <Form.Label>{t('预期用量')}</Form.Label>
              <Input
                placeholder={t('请描述您的预期月用量，例如：10万 tokens')}
                value={formValues.expected_volume}
                onChange={(value) => handleInputChange('expected_volume', value)}
                style={{ width: '100%' }}
              />
            </div>

            <div style={{ marginTop: '24px' }}>
              <Form.Label>{t('业务描述')}</Form.Label>
              <TextArea
                placeholder={t('请简单描述您的业务场景和计划')}
                value={formValues.business_description}
                onChange={(value) => handleInputChange('business_description', value)}
                style={{ width: '100%', minHeight: '120px' }}
              />
            </div>

            <div style={{ marginTop: '32px', textAlign: 'center' }}>
              <Button
                type='primary'
                size='large'
                theme='solid'
                loading={loading}
                onClick={handleSubmit}
                style={{ borderRadius: '50px', padding: '10px 48px', fontSize: '1rem', fontWeight: 600, minWidth: '200px' }}
              >
                {t('提交申请')}
              </Button>
            </div>
          </Form>
        </Card>

        <Card
          style={{
            borderRadius: '16px',
            marginTop: '24px',
            border: '1px solid var(--semi-color-border)',
          }}
          bodyStyle={{ padding: '32px' }}
        >
          <Title heading={4} style={{ fontSize: '1.2rem', fontWeight: 600, marginBottom: '16px' }}>
            {t('联系方式')}
          </Title>
          <Text style={{ fontSize: '1rem', color: 'var(--semi-color-text-1)', lineHeight: 1.8 }}>
            {t('如有疑问，请通过以下方式联系我们：')}
          </Text>
          <div style={{ marginTop: '16px' }}>
            <Text style={{ fontSize: '1rem' }}>Email: support@llmhub.ltd</Text>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default AffiliateApplication;