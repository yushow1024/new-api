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

import React, {
  useState,
  useContext,
  useCallback,
  useRef,
  useEffect,
} from 'react';
import { useTranslation } from 'react-i18next';
import {
  Typography,
  Button,
  Select,
  TextArea,
  InputNumber,
  Input,
  Toast,
  Spin,
  Card,
  Tag,
  Divider,
  Progress,
} from '@douyinfe/semi-ui';
import {
  Play,
  RefreshCw,
  Download,
  Film,
  Image,
  Copy,
  Shuffle,
} from 'lucide-react';
import { API, showError, showSuccess } from '../../helpers';
import { fetchTokenKeys } from '../../helpers/token';

const { Title, Text } = Typography;

const C = {
  bg: '#0a0a12',
  surface: '#12121f',
  surfaceHover: '#1a1a2e',
  border: 'rgba(100,116,255,0.15)',
  borderActive: 'rgba(100,116,255,0.45)',
  accent: '#6474ff',
  accentHover: '#7b88ff',
  accentLight: 'rgba(100,116,255,0.12)',
  text: '#e8e8f0',
  textSecondary: '#8888a0',
  textMuted: '#5a5a72',
  danger: '#ff4d6a',
  success: '#4ade80',
  warning: '#fbbf24',
  gradient: 'linear-gradient(135deg, #6474ff 0%, #a855f7 50%, #ff006e 100%)',
  gradientBg:
    'linear-gradient(135deg, rgba(100,116,255,0.08) 0%, rgba(168,85,247,0.08) 50%, rgba(255,0,110,0.08) 100%)',
};

const VIDEO_MODELS = [
  { value: 'sora-2-12s', label: 'Sora 2 (12s)', tag: 'OpenAI' },
  { value: 'veo_3_1', label: 'Veo 3.1', tag: 'Google' },
  { value: 'veo_3_1-hd-fl', label: 'Veo 3.1 HD FL', tag: 'Google' },
  { value: 'veo_3_1-hd', label: 'Veo 3.1 HD', tag: 'Google' },
  { value: 'omni_flash-10s', label: 'Omni Flash (10s)', tag: 'Gemini' },
  { value: 'grok-video-3', label: 'Grok Video 3', tag: 'xAI' },
];

const RESOLUTION_OPTIONS = [
  { value: 'square', label: '1:1 (1024×1024)', width: 1024, height: 1024 },
  { value: 'landscape', label: '16:9 (1280×720)', width: 1280, height: 720 },
  { value: 'portrait', label: '9:16 (720×1280)', width: 720, height: 1280 },
  { value: 'wide', label: '21:9 (1680×720)', width: 1680, height: 720 },
];

const VideoGeneration = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [generatedVideos, setGeneratedVideos] = useState([]);
  const [previewUrl, setPreviewUrl] = useState('');
  const [taskProgress, setTaskProgress] = useState(0);
  const [taskStatusText, setTaskStatusText] = useState('');
  const pollingRef = useRef(null);
  const tokenRef = useRef('');

  const [formValues, setFormValues] = useState({
    model: 'sora-2-12s',
    prompt: '',
    image: '',
    selectedResolution: 'landscape',
    duration: 8,
    width: 1280,
    height: 720,
    fps: 24,
    seed: '',
    n: 1,
    negative_prompt: '',
    quality_level: 'standard',
  });

  const handleInputChange = useCallback((field, value) => {
    setFormValues((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleResolutionChange = useCallback((value) => {
    const option = RESOLUTION_OPTIONS.find((r) => r.value === value);
    if (option) {
      setFormValues((prev) => ({
        ...prev,
        selectedResolution: value,
        width: option.width,
        height: option.height,
      }));
    }
  }, []);

  const generateSeed = useCallback(() => {
    const seed = Math.floor(Math.random() * 2147483647);
    handleInputChange('seed', seed);
  }, [handleInputChange]);

  useEffect(() => {
    return () => {
      if (pollingRef.current) {
        clearTimeout(pollingRef.current);
      }
    };
  }, []);

  const pollTask = useCallback(
    (taskId) => {
      if (!taskId || !tokenRef.current) return;

      pollingRef.current = setTimeout(async () => {
        try {
          const res = await API.get(`/v1/videos/${taskId}`, {
            headers: {
              Authorization: `Bearer ${tokenRef.current}`,
            },
          });
          const taskData = res.data;
          const status = taskData.status;
          const progress = taskData.progress || 0;

          setTaskProgress(Math.min(progress, 100));
          setTaskStatusText(getStatusText(status));

          if (status === 'completed' || status === 'succeeded') {
            const videoUrl =
              taskData.video_url || taskData.url || taskData.output_url || '';
            if (videoUrl) {
              setPreviewUrl(videoUrl);
            }
            setGenerating(false);
            showSuccess(t('视频生成完成'));
          } else if (status === 'failed' || status === 'cancelled') {
            const errMsg =
              taskData?.error?.message || t('视频生成失败，请重试');
            showError(errMsg);
            setGenerating(false);
          } else {
            pollTask(taskId);
          }
        } catch (error) {
          const msg =
            error?.response?.data?.error?.message ||
            error.message ||
            t('查询任务状态失败');
          showError(msg);
          setGenerating(false);
        }
      }, 2000);
    },
    [t],
  );

  const getStatusText = useCallback(
    (status) => {
      const map = {
        queued: t('排队中...'),
        pending: t('等待处理...'),
        processing: t('AI 正在生成视频...'),
        running: t('正在渲染...'),
      };
      return map[status] || t('处理中...');
    },
    [t],
  );

  const handleGenerate = async () => {
    Toast.info({ content: t('功能开发中'), duration: 3 });
    return
    if (!formValues.prompt.trim()) {
      showError(t('请输入视频描述'));
      return;
    }

    setGenerating(true);
    setGeneratedVideos([]);
    setPreviewUrl('');
    setTaskProgress(0);
    setTaskStatusText(t('排队中...'));

    try {
      const tokenKeys = await fetchTokenKeys();
      if (tokenKeys.length === 0) {
        showError(t('没有可用的令牌，请先创建令牌'));
        setGenerating(false);
        return;
      }
      tokenRef.current = tokenKeys[0];

      const formData = new FormData();
      formData.append('model', formValues.model);
      formData.append('prompt', formValues.prompt);

      if (formValues.image) {
        formData.append('image', formValues.image);
      }
      if (formValues.duration) {
        formData.append('duration', String(formValues.duration));
      }
      if (formValues.width) {
        formData.append('width', String(formValues.width));
      }
      if (formValues.height) {
        formData.append('height', String(formValues.height));
      }
      if (formValues.fps) {
        formData.append('fps', String(formValues.fps));
      }
      if (formValues.seed) {
        formData.append('seed', String(formValues.seed));
      }
      formData.append('n', String(formValues.n));
      formData.append('response_format', 'url');

      const metadata = {};
      if (formValues.negative_prompt) {
        metadata.negative_prompt = formValues.negative_prompt;
      }
      if (formValues.quality_level) {
        metadata.quality_level = formValues.quality_level;
      }
      if (Object.keys(metadata).length > 0) {
        formData.append('metadata', JSON.stringify(metadata));
      }

      const res = await API.post('/v1/videos', formData, {
        headers: {
          Authorization: `Bearer ${tokenKeys[0]}`,
        },
      });

      const data = res.data;
      const taskId = data.id;
      if (taskId) {
        pollTask(taskId);
      } else {
        showError(t('创建任务失败，未获取到任务 ID'));
        setGenerating(false);
      }
    } catch (error) {
      const msg =
        error?.response?.data?.error?.message ||
        error?.response?.data?.message ||
        error.message ||
        t('生成失败，请重试');
      showError(msg);
    } finally {
      setGenerating(false);
    }
  };

  const handleCopyPrompt = useCallback(() => {
    if (formValues.prompt) {
      navigator.clipboard.writeText(formValues.prompt);
      showSuccess(t('已复制'));
    }
  }, [formValues.prompt]);

  const selectStyle = {
    width: '100%',
    backgroundColor: C.surface,
    borderColor: C.border,
    color: C.text,
    borderRadius: '8px',
  };

  const inputStyle = {
    backgroundColor: C.surface,
    borderColor: C.border,
    color: C.text,
    borderRadius: '8px',
  };

  return (
    <div
      style={{
        height: 'calc(100vh - 64px)',
        display: 'flex',
        backgroundColor: C.bg,
        color: C.text,
        marginTop: '64px',
        overflow: 'hidden',
      }}
    >
      {/* 左侧参数面板 */}
      <div
        style={{
          width: '420px',
          minWidth: '420px',
          height: '100%',
          overflowY: 'auto',
          borderRight: `1px solid ${C.border}`,
          padding: '24px 20px',
          display: 'flex',
          flexDirection: 'column',
          gap: '20px',
        }}
        className='scrollbar-thin'
      >
        {/* 模型选择 */}
        <div>
          <Text
            style={{
              color: C.textSecondary,
              fontSize: '13px',
              fontWeight: 600,
              marginBottom: '8px',
              display: 'block',
            }}
          >
            {t('模型')}
          </Text>
          <Select
            value={formValues.model}
            onChange={(value) => handleInputChange('model', value)}
            style={selectStyle}
            optionList={VIDEO_MODELS.map((m) => ({
              value: m.value,
              label: (
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    width: '100%',
                  }}
                >
                  <span>{m.label}</span>
                  <Tag
                    size='small'
                    style={{
                      backgroundColor: C.accentLight,
                      color: C.accent,
                      border: 'none',
                      fontSize: '11px',
                    }}
                  >
                    {m.tag}
                  </Tag>
                </div>
              ),
            }))}
          />
        </div>

        {/* 提示词 */}
        <div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: '8px',
            }}
          >
            <Text
              style={{
                color: C.textSecondary,
                fontSize: '13px',
                fontWeight: 600,
              }}
            >
              {t('提示词')}
            </Text>
            <Button
              theme='borderless'
              size='small'
              icon={<Copy size={14} />}
              onClick={handleCopyPrompt}
              style={{ color: C.textMuted, fontSize: '12px' }}
            >
              {t('复制')}
            </Button>
          </div>
          <TextArea
            value={formValues.prompt}
            onChange={(value) => handleInputChange('prompt', value)}
            placeholder={t('请描述您想要生成的视频内容...')}
            style={{
              ...inputStyle,
              minHeight: '120px',
            }}
            maxLength={2000}
            showClear
          />
          <Text
            style={{
              color: C.textMuted,
              fontSize: '11px',
              marginTop: '4px',
              display: 'block',
              textAlign: 'right',
            }}
          >
            {(formValues.prompt || '').length}/2000
          </Text>
        </div>

        {/* 图片输入 */}
        <div>
          <Text
            style={{
              color: C.textSecondary,
              fontSize: '13px',
              fontWeight: 600,
              marginBottom: '8px',
              display: 'block',
            }}
          >
            {t('参考图片 (可选)')}
          </Text>
          <Input
            value={formValues.image}
            onChange={(value) => handleInputChange('image', value)}
            placeholder={t('输入图片 URL 或 Base64')}
            style={inputStyle}
            prefix={<Image size={16} />}
          />
        </div>

        {/* 分辨率 */}
        <div>
          <Text
            style={{
              color: C.textSecondary,
              fontSize: '13px',
              fontWeight: 600,
              marginBottom: '8px',
              display: 'block',
            }}
          >
            {t('分辨率')}
          </Text>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {RESOLUTION_OPTIONS.map((opt) => (
              <Tag
                key={opt.value}
                size='large'
                style={{
                  cursor: 'pointer',
                  backgroundColor:
                    formValues.selectedResolution === opt.value
                      ? C.accentLight
                      : C.surface,
                  color:
                    formValues.selectedResolution === opt.value
                      ? C.accent
                      : C.textSecondary,
                  border:
                    formValues.selectedResolution === opt.value
                      ? `1px solid ${C.accent}`
                      : `1px solid ${C.border}`,
                  borderRadius: '8px',
                  padding: '6px 12px',
                  transition: 'all 0.2s',
                }}
                onClick={() => handleResolutionChange(opt.value)}
              >
                {opt.label}
              </Tag>
            ))}
          </div>
        </div>

        {/* 时长 */}
        <div>
          <Text
            style={{
              color: C.textSecondary,
              fontSize: '13px',
              fontWeight: 600,
              marginBottom: '8px',
              display: 'block',
            }}
          >
            {t('视频时长')}
          </Text>
          <InputNumber
            value={formValues.duration}
            onChange={(value) => handleInputChange('duration', value)}
            min={1}
            max={120}
            step={1}
            suffix='s'
            placeholder={t('例如：8')}
            style={{ width: '100%', ...inputStyle }}
          />
        </div>

        <Divider style={{ borderColor: C.border }} />

        {/* 高级参数 */}
        <Text
          style={{ color: C.textSecondary, fontSize: '14px', fontWeight: 700 }}
        >
          {t('高级参数')}
        </Text>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '12px',
          }}
        >
          <div>
            <Text
              style={{
                color: C.textMuted,
                fontSize: '12px',
                marginBottom: '6px',
                display: 'block',
              }}
            >
              {t('宽度')}
            </Text>
            <InputNumber
              value={formValues.width}
              onChange={(value) => handleInputChange('width', value)}
              min={256}
              max={4096}
              step={64}
              style={inputStyle}
            />
          </div>
          <div>
            <Text
              style={{
                color: C.textMuted,
                fontSize: '12px',
                marginBottom: '6px',
                display: 'block',
              }}
            >
              {t('高度')}
            </Text>
            <InputNumber
              value={formValues.height}
              onChange={(value) => handleInputChange('height', value)}
              min={256}
              max={4096}
              step={64}
              style={inputStyle}
            />
          </div>
          <div>
            <Text
              style={{
                color: C.textMuted,
                fontSize: '12px',
                marginBottom: '6px',
                display: 'block',
              }}
            >
              {t('帧率')}
            </Text>
            <InputNumber
              value={formValues.fps}
              onChange={(value) => handleInputChange('fps', value)}
              min={1}
              max={60}
              style={inputStyle}
            />
          </div>
          <div>
            <Text
              style={{
                color: C.textMuted,
                fontSize: '12px',
                marginBottom: '6px',
                display: 'block',
              }}
            >
              {t('生成数量')}
            </Text>
            <InputNumber
              value={formValues.n}
              onChange={(value) => handleInputChange('n', value)}
              min={1}
              max={4}
              style={inputStyle}
            />
          </div>
          <div style={{ gridColumn: '1 / -1' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginBottom: '6px',
              }}
            >
              <Text style={{ color: C.textMuted, fontSize: '12px' }}>
                {t('随机种子')}
              </Text>
              <Button
                theme='borderless'
                size='small'
                icon={<Shuffle size={14} />}
                onClick={generateSeed}
                style={{ color: C.accent, fontSize: '11px', padding: 0 }}
              >
                {t('随机')}
              </Button>
            </div>
            <Input
              value={String(formValues.seed)}
              onChange={(value) => handleInputChange('seed', value)}
              placeholder={t('留空使用随机种子')}
              style={inputStyle}
            />
          </div>
        </div>

        {/* 负面提示词 */}
        <div>
          <Text
            style={{
              color: C.textMuted,
              fontSize: '12px',
              marginBottom: '6px',
              display: 'block',
            }}
          >
            {t('负面提示词 (可选)')}
          </Text>
          <Input
            value={formValues.negative_prompt}
            onChange={(value) => handleInputChange('negative_prompt', value)}
            placeholder={t('描述不想在视频中出现的内容')}
            style={inputStyle}
          />
        </div>

        {/* 生成按钮 */}
        <Button
          type='primary'
          theme='solid'
          size='large'
          loading={generating}
          onClick={handleGenerate}
          icon={<Play size={20} />}
          style={{
            width: '100%',
            borderRadius: '12px',
            height: '48px',
            fontSize: '16px',
            fontWeight: 700,
            background: C.gradient,
            border: 'none',
            marginTop: '8px',
          }}
        >
          {generating ? t('生成中...') : t('生成视频')}
        </Button>
      </div>

      {/* 右侧工作区 */}
      <div
        style={{
          flex: 1,
          height: '100%',
          overflowY: 'auto',
          padding: '32px',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: C.gradientBg,
        }}
        className='scrollbar-thin'
      >
        {generating ? (
          <div
            style={{ textAlign: 'center', maxWidth: '480px', width: '100%' }}
          >
            <div
              style={{
                width: '80px',
                height: '80px',
                borderRadius: '20px',
                background: C.gradient,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 24px',
              }}
            >
              <Film size={36} style={{ color: '#fff' }} />
            </div>
            <Text
              style={{
                color: C.textSecondary,
                display: 'block',
                fontSize: '15px',
                marginBottom: '6px',
              }}
            >
              {taskStatusText}
            </Text>
            <div style={{ marginTop: '20px', marginBottom: '12px' }}>
              <Progress
                percent={taskProgress}
                stroke={C.accent}
                style={{ flex: 1 }}
                showInfo={true}
                format={(percent) => `${percent}%`}
              />
            </div>
            <Text
              style={{
                color: C.textMuted,
                marginTop: '8px',
                display: 'block',
                fontSize: '13px',
              }}
            >
              {t('生成时间取决于视频时长和复杂度，请耐心等待')}
            </Text>
            <Button
              theme='borderless'
              type='danger'
              size='small'
              onClick={() => {
                if (pollingRef.current) {
                  clearTimeout(pollingRef.current);
                }
                setGenerating(false);
                setTaskProgress(0);
              }}
              style={{ marginTop: '16px', color: C.textMuted }}
            >
              {t('取消')}
            </Button>
          </div>
        ) : previewUrl ? (
          <div style={{ width: '100%', maxWidth: '800px' }}>
            <Card
              style={{
                backgroundColor: C.surface,
                borderColor: C.border,
                borderRadius: '16px',
                overflow: 'hidden',
              }}
              bodyStyle={{ padding: 0 }}
            >
              <video
                src={previewUrl}
                controls
                autoPlay
                loop
                style={{
                  width: '100%',
                  borderRadius: '16px 16px 0 0',
                  display: 'block',
                }}
              />
              <div
                style={{
                  padding: '16px 20px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <div>
                  <Text
                    style={{ color: C.text, fontSize: '14px', fontWeight: 600 }}
                  >
                    {t('生成的视频')}
                  </Text>
                  <Text
                    style={{
                      color: C.textMuted,
                      fontSize: '12px',
                      display: 'block',
                      marginTop: '2px',
                    }}
                  >
                    {formValues.width}×{formValues.height} ·{' '}
                    {formValues.duration}s · {formValues.fps}fps
                  </Text>
                </div>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <Button
                    theme='borderless'
                    icon={<RefreshCw size={16} />}
                    onClick={handleGenerate}
                    style={{ color: C.accent }}
                  >
                    {t('重新生成')}
                  </Button>
                  <Button
                    theme='borderless'
                    icon={<Download size={16} />}
                    onClick={() => window.open(previewUrl, '_blank')}
                    style={{ color: C.accent }}
                  >
                    {t('下载')}
                  </Button>
                </div>
              </div>
            </Card>

            {generatedVideos.length > 1 && (
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                  gap: '12px',
                  marginTop: '16px',
                }}
              >
                {generatedVideos.slice(1).map((video, idx) => (
                  <Card
                    key={idx}
                    style={{
                      backgroundColor: C.surface,
                      borderColor: C.border,
                      borderRadius: '12px',
                      overflow: 'hidden',
                      cursor: 'pointer',
                    }}
                    bodyStyle={{ padding: 0 }}
                    onClick={() => setPreviewUrl(video.url || video)}
                  >
                    <video
                      src={video.url || video}
                      style={{ width: '100%', display: 'block' }}
                      muted
                    />
                  </Card>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div style={{ textAlign: 'center', maxWidth: '480px' }}>
            <div
              style={{
                width: '80px',
                height: '80px',
                borderRadius: '20px',
                background: C.gradient,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 24px',
              }}
            >
              <Film size={36} style={{ color: '#fff' }} />
            </div>
            <Title
              heading={3}
              style={{ color: C.text, marginBottom: '8px', fontSize: '20px' }}
            >
              {t('视频生成')}
            </Title>
            <Text
              style={{
                color: C.textSecondary,
                fontSize: '14px',
                lineHeight: 1.6,
              }}
            >
              {t('在左侧输入提示词和参数，点击"生成视频"开始创建')}
            </Text>
            <div
              style={{
                marginTop: '32px',
                display: 'flex',
                gap: '16px',
                justifyContent: 'center',
                flexWrap: 'wrap',
              }}
            >
              {VIDEO_MODELS.slice(0, 3).map((m) => (
                <Tag
                  key={m.value}
                  style={{
                    backgroundColor: C.accentLight,
                    color: C.accent,
                    border: 'none',
                    borderRadius: '8px',
                    padding: '6px 14px',
                    fontSize: '13px',
                  }}
                >
                  {m.label}
                </Tag>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default VideoGeneration;
