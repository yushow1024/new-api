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
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  InputNumber,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { IconDelete, IconPlus } from '@douyinfe/semi-icons';
import { showSuccess } from '../../../../helpers';
import {
  createEmptyRuleGroup,
  buildRequestRuleExpr,
  tryParseRequestRuleExpr,
} from './requestRuleExpr';
import { RequestRuleGroupCard } from './RequestRuleGroup';

const { Text } = Typography;

// ---------------------------------------------------------------------------
// Per-second pricing editor
//
// Storage contract (matches backend billing_setting.GetPerSecondPrice /
// GetPerSecondRules exactly):
//   billing_setting.billing_per_second_price — model → $/sec number
//   billing_setting.billing_per_second_rules — model → rule-chain string of
//                                              `(cond ? X : 1)` factors
//                                              joined with ` * `
//
// Rules are persisted as a single string (mirroring tiered_expr's request
// rules shape); the visual editor renders one or more groups and the
// groups are flattened back to the string on every change.
// ---------------------------------------------------------------------------

export default function PerSecondPricingEditor({
  pricePerSec,
  onPricePerSecChange,
  ruleExpr,
  onRuleExprChange,
  t,
}) {
  // pricePerSec may arrive as string (from form state) or number; coerce.
  const priceNumber = useMemo(() => {
    const num = Number(pricePerSec);
    return Number.isFinite(num) ? num : 0;
  }, [pricePerSec]);

  // Detect parse failure BEFORE the empty-array fallback so the legacy
  // fallback block can render. tryParseRequestRuleExpr returns:
  //   - [] when input is empty
  //   - [...] when parsed cleanly
  //   - null when syntax is unrecognizable
  const parsedFailure = useMemo(() => {
    if (!ruleExpr || !ruleExpr.trim()) return false;
    return tryParseRequestRuleExpr(ruleExpr) === null;
  }, [ruleExpr]);

  // Parse the persisted rule string into an array of visual groups.
  const parsedGroups = useMemo(() => {
    if (!ruleExpr || !ruleExpr.trim()) return [];
    return tryParseRequestRuleExpr(ruleExpr) || [];
  }, [ruleExpr]);
  const [groups, setGroups] = useState(parsedGroups);

  // Re-sync local groups state when the persisted string changes externally
  // (e.g. user selected a different model in the list).
  useEffect(() => {
    setGroups(parsedGroups);
  }, [parsedGroups]);

  // Persist groups back to the joined `(g1) * (g2)` rule string.
  const persistGroups = useCallback(
    (nextGroups) => {
      setGroups(nextGroups);
      const expr = buildRequestRuleExpr(nextGroups);
      onRuleExprChange(expr || '');
    },
    [onRuleExprChange],
  );

  const handlePriceChange = (val) => {
    const num = Number(val);
    if (!Number.isFinite(num) || num < 0) return;
    onPricePerSecChange(num);
  };

  const copyPreview = () => {
    if (!savePreview) return;
    try {
      navigator.clipboard?.writeText(savePreview);
      showSuccess(t('已复制到剪贴板'));
    } catch {
      // ignore — clipboard may be unavailable
    }
  };

  // Save preview: combine base expression + all rule groups into a single
  // human-readable formula.
  const savePreview = useMemo(() => {
    const base = `seconds * ${priceNumber}`;
    const rules = (ruleExpr || '').trim();
    if (!rules) return base;
    return `(${base}) * ${rules}`;
  }, [priceNumber, ruleExpr]);

  // Current saved state mirrors preview
  const hasRules = Boolean((ruleExpr || '').trim());

  return (
    <div>
      {/* ─── Section 1: 固定价格 ($/秒) ─────────────────────────────── */}
      <Card
        bodyStyle={{ padding: 16 }}
        style={{ marginBottom: 12, background: 'var(--semi-color-fill-0)' }}
      >
        <div className='font-medium mb-2'>{t('固定价格')}</div>
        <Text
          size='small'
          type='secondary'
          style={{ display: 'block', marginBottom: 10 }}
        >
          {t('按秒单价。基础公式 = seconds × 单价；其它维度请在下方"请求条件调价"里加规则。')}
        </Text>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Text strong style={{ minWidth: 60 }}>{t('$/秒')}</Text>
          <InputNumber
            value={priceNumber}
            min={0}
            step={0.01}
            onChange={(val) => handlePriceChange(val ?? 0)}
            style={{ width: 200 }}
          />
          <Text size='small' type='secondary'>
            {t('基础公式：')}{' '}
            <code
              style={{
                background: 'var(--semi-color-fill-1)',
                padding: '2px 6px',
                borderRadius: 4,
              }}
            >
              seconds * {priceNumber}
            </code>
          </Text>
        </div>
      </Card>

      {/* ─── Section 2: 请求条件调价 ─────────────────────────────────── */}
      <Card
        bodyStyle={{ padding: 16 }}
        style={{ marginBottom: 12, background: 'var(--semi-color-fill-0)' }}
      >
        <div className='font-medium mb-2'>{t('请求条件调价')}</div>
        <Text type='secondary' size='small' style={{ display: 'block', marginBottom: 4 }}>
          {t('满足条件时，整单价格乘以 X；多条规则可同时叠加（依次相乘）。')}
        </Text>
        <Text type='secondary' size='small' style={{ display: 'block', marginBottom: 12 }}>
          {t('X 可以小于 1 当折扣用。需要"额外加固定费"或"只给某维度加价"，请用表达式/阶梯计费。')}
        </Text>

        {parsedFailure ? (
          // Legacy / manually edited expression can't be parsed into groups:
          // show a warning + raw textarea + "reset" button so the user can
          // start fresh without losing their saved value.
          <div
            style={{
              padding: '12px 16px',
              borderRadius: 8,
              border: '1px solid var(--semi-color-border)',
              background: 'var(--semi-color-bg-2)',
              marginBottom: 8,
            }}
          >
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 8,
              }}
            >
              <Tag color='grey' size='small'>
                {t('原始规则（无法解析为可视化结构）')}
              </Tag>
              <Button
                icon={<IconDelete />}
                type='danger'
                theme='borderless'
                size='small'
                onClick={() => persistGroups([])}
              >
                {t('清空并重新编辑')}
              </Button>
            </div>
            <TextArea
              value={ruleExpr}
              readOnly
              rows={2}
              style={{ width: '100%', fontFamily: 'monospace' }}
            />
          </div>
        ) : (
          <>
            {groups.map((group, gi) => (
              <RequestRuleGroupCard
                key={`rule-group-${gi}`}
                group={group}
                index={gi}
                t={t}
                onChange={(nextGroup) => {
                  const next = [...groups];
                  next[gi] = nextGroup;
                  persistGroups(next);
                }}
                onRemove={() => {
                  persistGroups(groups.filter((_, i) => i !== gi));
                }}
              />
            ))}
            <Button
              icon={<IconPlus />}
              size='small'
              theme='light'
              onClick={() => persistGroups([...groups, createEmptyRuleGroup()])}
              style={{ marginTop: 4 }}
            >
              {t('添加条件组')}
            </Button>
          </>
        )}
      </Card>

      {/* ─── Section 3: 保存预览 ──────────────────────────────────────── */}
      <Card
        bodyStyle={{ padding: 16 }}
        style={{ marginBottom: 12, background: 'var(--semi-color-fill-0)' }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 8,
          }}
        >
          <div className='font-medium'>{t('保存预览')}</div>
          <Button size='small' type='tertiary' onClick={copyPreview}>
            {t('复制')}
          </Button>
        </div>
        <Text
          size='small'
          type='secondary'
          style={{ display: 'block', marginBottom: 8 }}
        >
          {t('保存到后端的最终表达式：')}
        </Text>
        <div
          style={{
            background: 'var(--semi-color-fill-1)',
            borderRadius: 6,
            padding: '10px 12px',
            fontFamily: 'monospace',
            fontSize: 13,
            lineHeight: 1.6,
            wordBreak: 'break-all',
          }}
        >
          <div>
            <Tag color='violet' size='small' style={{ marginRight: 6 }}>
              base
            </Tag>
            seconds * {priceNumber}
          </div>
          {hasRules && (
            <div style={{ marginTop: 6 }}>
              <Tag color='blue' size='small' style={{ marginRight: 6 }}>
                rules
              </Tag>
              <span style={{ opacity: 0.7 }}>× </span>
              {ruleExpr}
            </div>
          )}
          {hasRules && (
            <div
              style={{
                marginTop: 8,
                paddingTop: 8,
                borderTop: '1px dashed var(--semi-color-border)',
              }}
            >
              <Tag color='green' size='small' style={{ marginRight: 6 }}>
                final
              </Tag>
              {savePreview}
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}