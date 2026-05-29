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

import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Typography, Button, Row, Col, Modal } from '@douyinfe/semi-ui';
import { IconPlusCircle } from '@douyinfe/semi-icons';

const { Title, Text, Paragraph } = Typography;

const C = {
  bg: '#06060F',
  surface: 'rgba(13,13,26,0.85)',
  border: 'rgba(139,92,246,0.15)',
  borderGlow: 'rgba(139,92,246,0.45)',
  neonPurple: '#8B5CF6',
  neonCyan: '#06D6FE',
  neonPink: '#FF006E',
  neonGreen: '#00FF88',
  text: '#F0F0FF',
  textSecondary: '#8892C0',
  textMuted: '#5A6480',
  gradient1: 'linear-gradient(135deg, #8B5CF6 0%, #06D6FE 50%, #FF006E 100%)',
  gradient2: 'linear-gradient(135deg, #8B5CF6 0%, #FF006E 100%)',
  gradient3: 'linear-gradient(135deg, #06D6FE 0%, #00FF88 100%)',
};

const HUDCorners = ({ color = C.neonPurple, size = 16 }) => (
  <>
    <span style={{ position: 'absolute', top: 0, left: 0, width: size, height: size, borderTop: `2px solid ${color}`, borderLeft: `2px solid ${color}`, opacity: 0.6 }} />
    <span style={{ position: 'absolute', top: 0, right: 0, width: size, height: size, borderTop: `2px solid ${color}`, borderRight: `2px solid ${color}`, opacity: 0.6 }} />
    <span style={{ position: 'absolute', bottom: 0, left: 0, width: size, height: size, borderBottom: `2px solid ${color}`, borderLeft: `2px solid ${color}`, opacity: 0.6 }} />
    <span style={{ position: 'absolute', bottom: 0, right: 0, width: size, height: size, borderBottom: `2px solid ${color}`, borderRight: `2px solid ${color}`, opacity: 0.6 }} />
  </>
);

const glow = (color, intensity = 10) => ({
  textShadow: `0 0 ${intensity}px ${color}, 0 0 ${intensity * 2}px ${color}`,
});

const Affiliate = () => {
  const { t } = useTranslation();
  const [showQR, setShowQR] = useState(false);

  return (
    <>
      <style>{`
        @keyframes cyberPulse { 0%, 100% { opacity: 0.4; } 50% { opacity: 0.8; } }
        @keyframes borderGlowPulse { 0%, 100% { border-color: rgba(139,92,246,0.2); box-shadow: 0 0 15px rgba(139,92,246,0.08); } 50% { border-color: rgba(139,92,246,0.5); box-shadow: 0 0 28px rgba(139,92,246,0.18); } }
        @keyframes floatUp { 0% { transform: translateY(0); opacity: 0; } 100% { transform: translateY(-100px); opacity: 0.25; } }
        .c-card { position:relative; background:rgba(13,13,26,0.85); backdrop-filter:blur(16px); border:1px solid rgba(139,92,246,0.15); border-radius:8px; transition:all 0.3s ease; overflow:hidden; }
        .c-card:hover { border-color:rgba(139,92,246,0.45); box-shadow:0 0 32px rgba(139,92,246,0.12), inset 0 0 32px rgba(139,92,246,0.02); transform:translateY(-2px); }
        .c-card::before { content:''; position:absolute; top:0; left:0; right:0; height:1px; background:linear-gradient(90deg, transparent, rgba(139,92,246,0.35), transparent); }
        .neon-line { height:1px; background:linear-gradient(90deg, transparent, rgba(139,92,246,0.35), transparent); margin:32px 0; border:none; }
        .scanline { position:absolute; top:0; left:0; right:0; bottom:0; background:repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.025) 2px, rgba(0,0,0,0.025) 4px); pointer-events:none; z-index:2; }
      `}</style>

      {/* QR Code Modal */}
      <Modal
        visible={showQR}
        onCancel={() => setShowQR(false)}
        footer={null}
        width={420}
        bodyStyle={{
          background: C.surface,
          padding: 0,
          textAlign: 'center',
        }}
        style={{ background: 'transparent' }}
      >
        <div style={{ position: 'relative', padding: '40px 32px 32px', overflow: 'hidden' }}>
          <HUDCorners color={C.neonCyan} size={16} />
          <div className="scanline" />

          <div style={{ position: 'relative', zIndex: 1 }}>
            <div style={{
              fontFamily: "'Orbitron', monospace",
              fontSize: '0.7rem',
              fontWeight: 600,
              letterSpacing: '0.25em',
              color: C.neonCyan,
              textTransform: 'uppercase',
              marginBottom: '20px',
            }}>
              Scan to Apply
            </div>

            <div style={{
              width: '220px',
              height: '220px',
              margin: '0 auto 24px',
              background: 'rgba(255,255,255,0.05)',
              border: `1px solid ${C.borderGlow}`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: '4px',
              position: 'relative',
            }}>
              <HUDCorners color={C.neonPurple} size={10} />
              {/* 替换为实际的微信二维码图片 */}
              <img
                src="/wechat-qr.png"
                alt="WeChat QR"
                style={{ width: '200px', height: '200px', objectFit: 'contain' }}
                onError={(e) => {
                  e.target.style.display = 'none';
                  e.target.nextSibling.style.display = 'flex';
                }}
              />
              <div style={{
                display: 'none',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '8px',
                color: C.textMuted,
                fontFamily: "'Exo 2', sans-serif",
                fontSize: '0.85rem',
              }}>
                <span style={{ fontSize: '2rem' }}>📱</span>
                <span>{t('请上传微信二维码')}</span>
                <span style={{ fontSize: '0.75rem' }}>/wechat-qr.png</span>
              </div>
            </div>

            <Title heading={4} style={{
              color: C.text,
              fontSize: '1.1rem',
              fontWeight: 600,
              marginBottom: '8px',
              fontFamily: "'Space Grotesk', sans-serif",
            }}>
              {t('微信扫码申请合作')}
            </Title>
            <Text style={{
              color: C.textSecondary,
              fontSize: '0.9rem',
              fontFamily: "'Exo 2', sans-serif",
              display: 'block',
              lineHeight: 1.6,
            }}>
              {t('扫码添加商务微信，获取独立站部署方案与合作详情')}
            </Text>
          </div>
        </div>
      </Modal>

      <div style={{ width: '100%', overflowX: 'hidden', backgroundColor: C.bg, color: C.text, minHeight: '100vh', fontFamily: "'Exo 2', sans-serif", marginTop: '64px' }}>
        {/* Background grid */}
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundImage: `linear-gradient(rgba(139,92,246,0.025) 1px, transparent 1px), linear-gradient(90deg, rgba(139,92,246,0.025) 1px, transparent 1px)`, backgroundSize: '60px 60px', pointerEvents: 'none', zIndex: 0 }} />
        {/* Ambient orbs */}
        <div style={{ position: 'fixed', width: '600px', height: '600px', borderRadius: '50%', background: 'radial-gradient(circle, rgba(139,92,246,0.07) 0%, transparent 70%)', top: '-200px', right: '-200px', pointerEvents: 'none', zIndex: 0, filter: 'blur(80px)' }} />
        <div style={{ position: 'fixed', width: '400px', height: '400px', borderRadius: '50%', background: 'radial-gradient(circle, rgba(6,214,254,0.05) 0%, transparent 70%)', bottom: '-100px', left: '-100px', pointerEvents: 'none', zIndex: 0, filter: 'blur(80px)' }} />

        {/* ===== HERO ===== */}
        <div style={{ position: 'relative', padding: '100px 20px 80px', textAlign: 'center', overflow: 'hidden', zIndex: 1 }}>
          <div className="scanline" />
          <span style={{ position: 'absolute', top: '15%', left: '5%', width: '120px', height: '1px', background: `linear-gradient(90deg, ${C.neonCyan}, transparent)`, opacity: 0.35 }} />
          <span style={{ position: 'absolute', top: '15%', left: '5%', width: '1px', height: '60px', background: `linear-gradient(180deg, ${C.neonCyan}, transparent)`, opacity: 0.35 }} />
          <span style={{ position: 'absolute', top: '15%', right: '5%', width: '120px', height: '1px', background: `linear-gradient(270deg, ${C.neonPink}, transparent)`, opacity: 0.35 }} />
          <span style={{ position: 'absolute', top: '15%', right: '5%', width: '1px', height: '60px', background: `linear-gradient(180deg, ${C.neonPink}, transparent)`, opacity: 0.35 }} />

          <div style={{ display: 'inline-block', padding: '6px 20px', marginBottom: '32px', border: `1px solid ${C.borderGlow}`, background: 'rgba(139,92,246,0.07)', fontFamily: "'Orbitron', monospace", fontSize: '0.72rem', fontWeight: 600, letterSpacing: '0.22em', color: C.neonPurple, textTransform: 'uppercase', animation: 'borderGlowPulse 3s ease-in-out infinite' }}>
            AI API Partnership Program
          </div>

          <Title heading={1} style={{ color: C.text, fontSize: 'clamp(2.2rem, 6vw, 4rem)', fontWeight: 800, marginBottom: '20px', letterSpacing: '-0.02em', fontFamily: "'Space Grotesk', sans-serif", lineHeight: 1.15, ...glow(C.neonPurple, 6) }}>
            {t('代理合作')}
          </Title>

          <Title heading={3} style={{ color: C.textSecondary, fontWeight: 400, fontSize: 'clamp(1rem, 2.5vw, 1.3rem)', maxWidth: '700px', margin: '0 auto 20px', fontFamily: "'Exo 2', sans-serif", letterSpacing: '0.02em', lineHeight: 1.6 }}>
            {t('独立建站')}
          </Title>

          <div style={{ width: '80px', height: '2px', background: C.gradient1, margin: '32px auto', borderRadius: '1px', boxShadow: `0 0 12px ${C.neonPurple}` }} />

          <div style={{ display: 'inline-block', background: 'rgba(139,92,246,0.1)', backdropFilter: 'blur(10px)', border: `1px solid ${C.borderGlow}`, padding: '10px 32px', fontSize: '1.1rem', fontWeight: 700, fontFamily: "'Orbitron', monospace", letterSpacing: '0.05em', color: C.neonCyan, ...glow(C.neonCyan, 4) }}>
            {t('首充 30% 分成')}
          </div>

          {['01', '10', '11', '00'].map((bin, i) => (
            <span key={bin} style={{ position: 'absolute', left: `${20 + i * 22}%`, top: `${30 + (i % 2) * 40}%`, fontFamily: "'Orbitron', monospace", fontSize: '0.65rem', color: `rgba(6,214,254,${0.06 + i * 0.015})`, pointerEvents: 'none', animation: `floatUp ${4 + i}s ease-in-out infinite` }}>{bin}</span>
          ))}
        </div>

        {/* ===== MAIN CONTENT ===== */}
        <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 20px 80px', position: 'relative', zIndex: 1 }}>
          {/* Intro */}
          <div style={{ textAlign: 'center', marginBottom: '80px' }}>
            <Title heading={2} style={{ fontSize: 'clamp(1.5rem, 4vw, 2.2rem)', fontWeight: 700, marginBottom: '24px', fontFamily: "'Space Grotesk', sans-serif", color: C.text, letterSpacing: '-0.01em' }}>
              {t('AI 接口代理合作，支持邀请分成与独立站点搭建')}
            </Title>
            <Text style={{ fontSize: '1.05rem', lineHeight: 1.8, maxWidth: '800px', margin: '0 auto', color: C.textSecondary, fontFamily: "'Exo 2', sans-serif", display: 'block' }}>
              {t('面向想开展 AI API 业务的合作伙伴，平台提供渠道资源、独立站部署、证书配置、倍率初始化、模型监控、服务器监控和后续技术协助。你专注获客和运营，系统交付与基础运维由技术支持完成。')}
            </Text>
          </div>

          {/* ===== PRICING CARD ===== */}
          <div className="c-card" style={{ marginBottom: '80px', padding: '48px 40px' }}>
            <HUDCorners color={C.neonCyan} size={20} />
            <Title heading={3} style={{ fontSize: '1.2rem', fontWeight: 600, textAlign: 'center', marginBottom: '36px', fontFamily: "'Space Grotesk', sans-serif", color: C.textSecondary, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
              {t('综合拿货测算')}
            </Title>
            <div style={{ textAlign: 'center', marginBottom: '40px' }}>
              <div style={{ fontSize: 'clamp(2.5rem, 8vw, 5rem)', fontWeight: 900, fontFamily: "'Orbitron', monospace", letterSpacing: '0.03em', background: C.gradient1, WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text', filter: `drop-shadow(0 0 20px ${C.neonPurple})` }}>
                {t('约 0.3 倍率')}
              </div>
            </div>
            <Text style={{ textAlign: 'center', fontSize: '0.95rem', display: 'block', marginBottom: '36px', color: C.textSecondary, fontFamily: "'Exo 2', sans-serif", lineHeight: 1.7, maxWidth: '650px', margin: '0 auto 36px' }}>
              {t('客户按 0.4 倍率使用；合作伙伴充值 100 实际到账 133，折算后综合拿货成本约为 0.3 倍率。')}
            </Text>

            <Row gutter={[24, 24]} style={{ marginBottom: '36px' }}>
              {[
                { value: '100', label: t('充值金额'), color: C.neonCyan },
                { value: '133', label: t('实际到账'), color: C.neonGreen },
                { value: '0.4', label: t('客户倍率'), color: C.neonPink },
              ].map((item, i) => (
                <Col span={8} key={i}>
                  <div style={{ textAlign: 'center', padding: '20px 12px', background: 'rgba(139,92,246,0.03)', borderRadius: '6px', border: '1px solid rgba(139,92,246,0.08)' }}>
                    <div style={{ fontSize: '2.5rem', fontWeight: 800, fontFamily: "'Orbitron', monospace", letterSpacing: '0.05em', color: item.color }}>{item.value}</div>
                    <Text style={{ fontSize: '0.88rem', color: C.textMuted, fontFamily: "'Exo 2', sans-serif", marginTop: '8px', display: 'block' }}>{item.label}</Text>
                  </div>
                </Col>
              ))}
            </Row>

            <div className="neon-line" />

            <div style={{ background: 'rgba(6,214,254,0.03)', borderRadius: '6px', padding: '28px', textAlign: 'center', border: '1px solid rgba(6,214,254,0.06)' }}>
              <Text style={{ fontSize: '0.85rem', color: C.textMuted, fontFamily: "'Exo 2', sans-serif", display: 'block', marginBottom: '12px' }}>{t('成本换算')}</Text>
              <div style={{ fontSize: '1.1rem', fontWeight: 600, fontFamily: "'Orbitron', monospace", color: C.neonCyan, letterSpacing: '0.03em' }}>0.4 × 100 / 133 ≈ 0.3</div>
              <Text style={{ fontSize: '0.85rem', color: C.textMuted, marginTop: '12px', display: 'block', fontFamily: "'Exo 2', sans-serif", lineHeight: 1.6 }}>
                {t('按官方美元计价折算：0.3 / 7 ≈ 4.3%。独立站点可设置高于 0.4 的销售倍率，差额就是利润空间。')}
              </Text>
            </div>
          </div>

          {/* ===== TWO MODES ===== */}
          <Row gutter={[32, 32]} style={{ marginBottom: '80px' }}>
            {[
              { badge: t('轻量开始'), grad: C.gradient2, title: t('模式一：邀请注册，按消费分成'), desc: t('无需自建站点，直接邀请用户到本站注册充值。用户首次充值消费后，可获得 30% 分成；后续充值消费按 15% 分成，后台可查看邀请充值、消费和返佣记录。'), accent: C.neonPink },
              { badge: t('长期运营'), grad: C.gradient3, title: t('模式二：独立站点，自己定价'), desc: t('适合希望沉淀自有品牌和客户的人。你拥有独立域名、后台、用户体系和价格策略，平台提供渠道资源、部署初始化、监控配置和后续技术协助。'), accent: C.neonCyan },
            ].map((mode, i) => (
              <Col span={12} key={i}>
                <div className="c-card" style={{ padding: '36px 32px', height: '100%' }}>
                  <HUDCorners color={mode.accent} size={12} />
                  <div style={{ display: 'inline-block', background: mode.grad, color: '#fff', padding: '5px 14px', fontSize: '0.76rem', fontWeight: 700, fontFamily: "'Orbitron', monospace", letterSpacing: '0.1em', textTransform: 'uppercase', marginBottom: '24px' }}>{mode.badge}</div>
                  <Title heading={4} style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '16px', fontFamily: "'Space Grotesk', sans-serif", color: C.text, letterSpacing: '-0.01em', lineHeight: 1.4 }}>{mode.title}</Title>
                  <Paragraph style={{ color: C.textSecondary, lineHeight: 1.8, fontSize: '0.95rem', fontFamily: "'Exo 2', sans-serif", margin: 0 }}>{mode.desc}</Paragraph>
                </div>
              </Col>
            ))}
          </Row>

          {/* ===== DELIVERY ===== */}
          <div style={{ marginBottom: '80px' }}>
            <div style={{ textAlign: 'center', marginBottom: '48px' }}>
              <div style={{ fontFamily: "'Orbitron', monospace", fontSize: '0.72rem', fontWeight: 600, letterSpacing: '0.3em', color: C.neonCyan, textTransform: 'uppercase', marginBottom: '16px' }}>Delivery</div>
              <Title heading={2} style={{ fontSize: 'clamp(1.5rem, 4vw, 2rem)', fontWeight: 700, marginBottom: '16px', fontFamily: "'Space Grotesk', sans-serif", color: C.text }}>{t('免费部署包含什么')}</Title>
              <Text style={{ color: C.textSecondary, fontSize: '1rem', display: 'block', maxWidth: '700px', margin: '0 auto', fontFamily: "'Exo 2', sans-serif", lineHeight: 1.6 }}>{t('部署不是只把程序跑起来，而是把可运营所需的访问、证书、服务、倍率、监控和健康检查一起处理好。')}</Text>
            </div>
            <Row gutter={[16, 16]}>
              {[
                t('部署基于开源 new-api 深度开发的系统'), t('提供可用二级域名'), t('申请并配置 SSL 证书'), t('配置 Nginx HTTPS 代理转发'), t('安装 Docker / MySQL / Redis'),
                t('初始化后台和接入上游渠道'), t('同步模型倍率与默认分组配置'), t('配置服务器监控面板'), t('配置模型可用性监控面板'), t('完成接口健康检查和公网访问验证'),
              ].map((item, idx) => (
                <Col span={12} key={idx}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '14px', padding: '18px 20px', background: 'rgba(13,13,26,0.6)', borderRadius: '4px', border: '1px solid rgba(139,92,246,0.1)', transition: 'all 0.25s ease' }}
                    onMouseEnter={e => { e.currentTarget.style.borderColor = C.borderGlow; e.currentTarget.style.boxShadow = '0 0 20px rgba(139,92,246,0.08)'; }}
                    onMouseLeave={e => { e.currentTarget.style.borderColor = 'rgba(139,92,246,0.1)'; e.currentTarget.style.boxShadow = 'none'; }}>
                    <div style={{ width: '32px', height: '32px', display: 'flex', alignItems: 'center', justifyContent: 'center', border: `1px solid ${C.borderGlow}`, fontFamily: "'Orbitron', monospace", fontSize: '0.78rem', fontWeight: 700, color: C.neonCyan, flexShrink: 0, background: 'rgba(6,214,254,0.04)' }}>{(idx + 1).toString().padStart(2, '0')}</div>
                    <Text style={{ fontSize: '0.92rem', color: C.textSecondary, fontFamily: "'Exo 2', sans-serif" }}>{item}</Text>
                  </div>
                </Col>
              ))}
            </Row>
          </div>

          {/* ===== OPERATION ===== */}
          <div style={{ marginBottom: '80px' }}>
            <div style={{ textAlign: 'center', marginBottom: '48px' }}>
              <div style={{ fontFamily: "'Orbitron', monospace", fontSize: '0.72rem', fontWeight: 600, letterSpacing: '0.3em', color: C.neonPink, textTransform: 'uppercase', marginBottom: '16px' }}>Operation</div>
              <Title heading={2} style={{ fontSize: 'clamp(1.5rem, 4vw, 2rem)', fontWeight: 700, marginBottom: '16px', fontFamily: "'Space Grotesk', sans-serif", color: C.text }}>{t('独立站点的运营能力')}</Title>
              <Text style={{ color: C.textSecondary, fontSize: '1rem', display: 'block', maxWidth: '700px', margin: '0 auto', fontFamily: "'Exo 2', sans-serif", lineHeight: 1.6 }}>{t('系统不只是转发接口，还包含面向长期运营的监控、调度、数据统计和渠道管理能力。')}</Text>
            </div>
            <Row gutter={[20, 20]}>
              {[
                { title: t('服务器监控'), desc: t('实时查看 CPU、内存、磁盘、网络 IO 等运行状态，便于判断服务压力。'), id: '01' },
                { title: t('模型可用性监控'), desc: t('查看模型调用成功率、调用耗时、异常状态和可用趋势。'), id: '02' },
                { title: t('智能调度'), desc: t('根据渠道和模型状态动态分发请求，降低单点异常对用户调用的影响。'), id: '03' },
                { title: t('异常模型处理'), desc: t('可对异常模型进行禁用、开启或调度调整，恢复后再重新放量。'), id: '04' },
                { title: t('经营数据看板'), desc: t('支持余额、额度、模型消耗、调用日志、用户统计等运营数据查看。'), id: '05' },
                { title: t('渠道与倍率管理'), desc: t('支持渠道管理、分组倍率、模型倍率、日志追踪和后续扩展。'), id: '06' },
              ].map((item, idx) => (
                <Col span={12} key={idx}>
                  <div className="c-card" style={{ padding: '28px', height: '100%' }}>
                    <div style={{ fontFamily: "'Orbitron', monospace", fontSize: '0.68rem', color: C.neonPurple, marginBottom: '16px', letterSpacing: '0.12em' }}>{item.id}</div>
                    <Title heading={5} style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '12px', fontFamily: "'Space Grotesk', sans-serif", color: C.text }}>{item.title}</Title>
                    <Text style={{ fontSize: '0.9rem', lineHeight: 1.7, color: C.textSecondary, fontFamily: "'Exo 2', sans-serif" }}>{item.desc}</Text>
                  </div>
                </Col>
              ))}
            </Row>
          </div>

          {/* ===== SERVICES ===== */}
          <div style={{ marginBottom: '80px' }}>
            <div style={{ textAlign: 'center', marginBottom: '48px' }}>
              <div style={{ fontFamily: "'Orbitron', monospace", fontSize: '0.72rem', fontWeight: 600, letterSpacing: '0.3em', color: C.neonGreen, textTransform: 'uppercase', marginBottom: '16px' }}>Services</div>
              <Title heading={2} style={{ fontSize: 'clamp(1.5rem, 4vw, 2rem)', fontWeight: 700, marginBottom: '16px', fontFamily: "'Space Grotesk', sans-serif", color: C.text }}>{t('还能提供的配套服务')}</Title>
              <Text style={{ color: C.textSecondary, fontSize: '1rem', display: 'block', maxWidth: '700px', margin: '0 auto', fontFamily: "'Exo 2', sans-serif", lineHeight: 1.6 }}>{t('从部署交付到服务器、CDN、证书、倍率和后续问题处理，尽量减少合作伙伴的基础运维成本。')}</Text>
            </div>
            <Row gutter={[20, 20]}>
              {[
                { title: t('整站交付'), desc: t('服务器环境、容器服务、数据库、缓存、反向代理和后台初始化一次性配置完成。') },
                { title: t('HTTPS 与证书'), desc: t('域名绑定后申请证书，配置 Nginx HTTPS 代理，并验证公网访问。') },
                { title: t('倍率初始化'), desc: t('同步模型倍率，配置默认分组倍率，站点后台可继续独立调整利润空间。') },
                { title: t('监控面板'), desc: t('配置服务器监控和模型可用性监控，持续观察调用质量和系统压力。') },
                { title: t('CDN 加速'), desc: t('可协助接入 CDN，适合美国节点服务器，降低跨境访问延迟。') },
                { title: t('服务器折扣'), desc: t('通过推荐渠道购买服务器可享 75 折，代购和配置收 20 元手续费。') },
              ].map((item, idx) => (
                <Col span={12} key={idx}>
                  <div className="c-card" style={{ padding: '28px', height: '100%' }}>
                    <div style={{ width: '8px', height: '8px', background: C.neonGreen, boxShadow: `0 0 8px ${C.neonGreen}`, marginBottom: '16px' }} />
                    <Title heading={5} style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '12px', fontFamily: "'Space Grotesk', sans-serif", color: C.text }}>{item.title}</Title>
                    <Text style={{ fontSize: '0.9rem', lineHeight: 1.7, color: C.textSecondary, fontFamily: "'Exo 2', sans-serif" }}>{item.desc}</Text>
                  </div>
                </Col>
              ))}
            </Row>
          </div>

          {/* ===== TECH BANNER ===== */}
          <div style={{ position: 'relative', background: 'rgba(13,13,26,0.9)', border: '1px solid rgba(139,92,246,0.2)', padding: '48px 40px', marginBottom: '80px', textAlign: 'center', overflow: 'hidden' }}>
            <HUDCorners color={C.neonPurple} size={24} />
            <div className="scanline" />
            <div style={{ fontFamily: "'Orbitron', monospace", fontSize: '0.68rem', fontWeight: 600, letterSpacing: '0.28em', color: C.neonCyan, textTransform: 'uppercase', marginBottom: '20px', position: 'relative', zIndex: 1 }}>Powered by Open Source</div>
            <Title heading={3} style={{ color: C.text, fontSize: '1.4rem', fontWeight: 700, marginBottom: '20px', fontFamily: "'Space Grotesk', sans-serif", position: 'relative', zIndex: 1 }}>{t('为什么基于 new-api 深度开发')}</Title>
            <Paragraph style={{ color: C.textSecondary, fontSize: '0.95rem', lineHeight: 1.8, maxWidth: '750px', margin: '0 auto', fontFamily: "'Exo 2', sans-serif", position: 'relative', zIndex: 1 }}>
              {t('系统基于开源 new-api 项目深度开发，优势是成熟稳定、功能完整、可控性强，支持用户管理、余额充值、额度统计、渠道管理、模型倍率设置、日志查询、数据看板、服务器监控、模型可用性监控和动态请求分发等功能。相比从零开发，部署更快、成本更低，后续也方便继续定制和扩展。')}
            </Paragraph>
            <div style={{ position: 'absolute', bottom: 0, right: 0, width: '80px', height: '80px', background: `linear-gradient(135deg, transparent 50%, rgba(139,92,246,0.05) 50%)`, pointerEvents: 'none' }} />
          </div>

          {/* ===== CTA ===== */}
          <div style={{ textAlign: 'center' }}>
            <Button type='primary' size='large' theme='solid'
              onClick={() => setShowQR(true)}
              style={{ borderRadius: '2px', padding: '14px 56px', fontSize: '1.05rem', fontWeight: 700, fontFamily: "'Orbitron', monospace", letterSpacing: '0.08em', background: C.gradient2, border: 'none', boxShadow: `0 0 30px rgba(139,92,246,0.4), 0 0 60px rgba(255,0,110,0.15)`, transition: 'all 0.3s ease' }}
              icon={<IconPlusCircle />}
              onMouseEnter={e => { e.currentTarget.style.boxShadow = `0 0 40px rgba(139,92,246,0.6), 0 0 80px rgba(255,0,110,0.25)`; e.currentTarget.style.transform = 'translateY(-2px)'; }}
              onMouseLeave={e => { e.currentTarget.style.boxShadow = `0 0 30px rgba(139,92,246,0.4), 0 0 60px rgba(255,0,110,0.15)`; e.currentTarget.style.transform = 'translateY(0)'; }}
            >
              {t('立即申请合作')}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
};

export default Affiliate;