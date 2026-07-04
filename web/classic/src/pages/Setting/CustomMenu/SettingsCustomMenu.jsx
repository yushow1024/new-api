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

import React, { useEffect, useRef, useState } from 'react';
import {
  Banner,
  Button,
  Divider,
  Form,
  Input,
  Modal,
  Radio,
  RadioGroup,
  Space,
  Spin,
  Table,
} from '@douyinfe/semi-ui';
import {
  IconDelete,
  IconEdit,
  IconPlus,
  IconSaveStroked,
  IconSearch,
} from '@douyinfe/semi-icons';
import {
  API,
  compareObjects,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { getCustomMenuIcon } from '../../../helpers/render';
import { useTranslation } from 'react-i18next';

/**
 * 自定义菜单配置面板。
 * 数据形态(后端 setting.CustomMenu):
 *   [{ name: string, url: string, icon: string }, ...]
 *
 * 与「聊天设置」不同的地方:
 *   - 数据结构是固定字段对象,不是 {name: url} 的 map
 *   - 多一个 icon 字段(lucide-react 图标名,例如 "Bot"、"Sparkles")
 *
 * 与「聊天设置」相同的地方:
 *   - 通过 PUT /api/option/  key=CustomMenu 落盘
 *   - 提供可视化 + JSON 双模式编辑
 */
export default function SettingsCustomMenu(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({ CustomMenu: '[]' });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);
  const [editMode, setEditMode] = useState('visual');

  const [menuItems, setMenuItems] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState(null);
  const [isEdit, setIsEdit] = useState(false);
  const [searchText, setSearchText] = useState('');
  const modalFormRef = useRef();

  const jsonToItems = (jsonString) => {
    try {
      const arr = JSON.parse(jsonString);
      if (!Array.isArray(arr)) return [];
      return arr.map((it, index) => ({
        id: index,
        name: it?.name || '',
        url: it?.url || '',
        icon: it?.icon || '',
      }));
    } catch (e) {
      console.error('CustomMenu JSON parse error:', e);
      return [];
    }
  };

  const itemsToJson = (items) =>
    JSON.stringify(
      items.map(({ name, url, icon }) => ({ name, url, icon })),
      null,
      2,
    );

  const syncJsonToItems = () => {
    setMenuItems(jsonToItems(inputs.CustomMenu));
  };

  const syncItemsToJson = (items) => {
    const json = itemsToJson(items);
    setInputs((prev) => ({ ...prev, CustomMenu: json }));
    if (refForm.current && editMode === 'json') {
      refForm.current.setValues({ CustomMenu: json });
    }
  };

  async function onSubmit() {
    try {
      if (editMode === 'json' && refForm.current) {
        try {
          await refForm.current.validate();
        } catch (err) {
          console.error('Validation failed:', err);
          showError(t('请检查输入'));
          return;
        }
      }
      const updateArray = compareObjects(inputs, inputsRow);
      if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
      const requestQueue = updateArray.map((item) =>
        API.put('/api/option/', { key: item.key, value: inputs[item.key] }),
      );
      setLoading(true);
      try {
        const res = await Promise.all(requestQueue);
        if (res.includes(undefined)) {
          if (requestQueue.length > 1) {
            showError(t('部分保存失败，请重试'));
          }
          return;
        }
        showSuccess(t('保存成功'));
        props.refresh();
      } catch {
        showError(t('保存失败，请重试'));
      } finally {
        setLoading(false);
      }
    } catch (e) {
      showError(t('请检查输入'));
      console.error(e);
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        if (key === 'CustomMenu') {
          // 后端返回的字符串保险起见 normalize 一遍
          try {
            const obj = JSON.parse(props.options[key] || '[]');
            currentInputs[key] = JSON.stringify(obj, null, 2);
          } catch {
            currentInputs[key] = '[]';
          }
        } else {
          currentInputs[key] = props.options[key];
        }
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    if (refForm.current) refForm.current.setValues(currentInputs);
    setMenuItems(jsonToItems(currentInputs.CustomMenu || '[]'));
  }, [props.options]);

  useEffect(() => {
    if (editMode === 'visual') syncJsonToItems();
  }, [inputs.CustomMenu, editMode]);

  useEffect(() => {
    if (refForm.current && editMode === 'json') {
      refForm.current.setValues(inputs);
    }
  }, [editMode, inputs]);

  const handleAdd = () => {
    setEditing({ name: '', url: '', icon: '' });
    setIsEdit(false);
    setModalVisible(true);
    setTimeout(() => {
      if (modalFormRef.current) {
        modalFormRef.current.setValues({ name: '', url: '', icon: '' });
      }
    }, 100);
  };

  const handleEdit = (record) => {
    setEditing({ ...record });
    setIsEdit(true);
    setModalVisible(true);
    setTimeout(() => {
      if (modalFormRef.current) {
        modalFormRef.current.setValues(record);
      }
    }, 100);
  };

  const handleDelete = (id) => {
    const next = menuItems.filter((it) => it.id !== id);
    setMenuItems(next);
    syncItemsToJson(next);
    showSuccess(t('删除成功'));
  };

  const handleModalOk = () => {
    if (!modalFormRef.current) return;
    modalFormRef.current
      .validate()
      .then((values) => {
        // 名称重复校验(同 ChatsSetting)
        const duplicated = menuItems.some(
          (it) => it.name === values.name && (!isEdit || it.id !== editing.id),
        );
        if (duplicated) {
          showError(t('菜单名称已存在，请使用其他名称'));
          return;
        }
        let next;
        if (isEdit) {
          next = menuItems.map((it) =>
            it.id === editing.id
              ? {
                  ...editing,
                  name: values.name,
                  url: values.url,
                  icon: values.icon || '',
                }
              : it,
          );
        } else {
          const maxId =
            menuItems.length > 0
              ? Math.max(...menuItems.map((it) => it.id))
              : -1;
          next = [
            ...menuItems,
            {
              id: maxId + 1,
              name: values.name,
              url: values.url,
              icon: values.icon || '',
            },
          ];
        }
        setMenuItems(next);
        syncItemsToJson(next);
        setModalVisible(false);
        setEditing(null);
        showSuccess(isEdit ? t('编辑成功') : t('添加成功'));
      })
      .catch((err) => console.error('Modal form validation error:', err));
  };

  const handleModalCancel = () => {
    setModalVisible(false);
    setEditing(null);
  };

  const filtered = menuItems.filter(
    (it) =>
      !searchText || it.name.toLowerCase().includes(searchText.toLowerCase()),
  );

  const highlight = (text) => {
    if (!text) return text;
    const parts = text.split(/(\{address\}|\{key\})/g);
    return parts.map((part, idx) => {
      if (part === '{address}') {
        return (
          <span key={idx} style={{ color: '#0077cc', fontWeight: 600 }}>
            {part}
          </span>
        );
      } else if (part === '{key}') {
        return (
          <span key={idx} style={{ color: '#ff6b35', fontWeight: 600 }}>
            {part}
          </span>
        );
      }
      return part;
    });
  };

  const columns = [
    {
      title: t('图标'),
      dataIndex: 'icon',
      key: 'icon',
      width: 80,
      render: (icon) => (
        <span style={{ display: 'inline-flex', alignItems: 'center' }}>
          {getCustomMenuIcon(icon)}
        </span>
      ),
    },
    {
      title: t('菜单名称'),
      dataIndex: 'name',
      key: 'name',
      render: (text) => text || t('未命名'),
    },
    {
      title: t('URL链接'),
      dataIndex: 'url',
      key: 'url',
      render: (text) => (
        <div style={{ maxWidth: 320, wordBreak: 'break-all' }}>
          {highlight(text)}
        </div>
      ),
    },
    {
      title: t('操作'),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type='primary'
            icon={<IconEdit />}
            size='small'
            onClick={() => handleEdit(record)}
          >
            {t('编辑')}
          </Button>
          <Button
            type='danger'
            icon={<IconDelete />}
            size='small'
            onClick={() => handleDelete(record.id)}
          >
            {t('删除')}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Spin spinning={loading}>
      <Space vertical style={{ width: '100%' }}>
        <Form.Section text={t('自定义菜单')}>
          <Banner
            type='info'
            description={t(
              '在侧边栏“聊天”分组下方按顺序显示。链接中的 {key} 将被替换为当前 API 密钥，{address} 替换为系统设置的服务器地址。图标填写 lucide-react 图标名（例如 Bot、Sparkles、Wand2、BookOpen），留空则使用默认图标。',
            )}
          />

          <Divider />

          <div style={{ marginBottom: 16 }}>
            <span style={{ marginRight: 16, fontWeight: 600 }}>
              {t('编辑模式')}:
            </span>
            <RadioGroup
              type='button'
              value={editMode}
              onChange={(e) => {
                const newMode = e.target.value;
                setEditMode(newMode);
                setTimeout(() => {
                  if (newMode === 'json' && refForm.current) {
                    refForm.current.setValues(inputs);
                  }
                }, 100);
              }}
            >
              <Radio value='visual'>{t('可视化编辑')}</Radio>
              <Radio value='json'>{t('JSON编辑')}</Radio>
            </RadioGroup>
          </div>

          {editMode === 'visual' ? (
            <div>
              <Space style={{ marginBottom: 16 }}>
                <Button type='primary' icon={<IconPlus />} onClick={handleAdd}>
                  {t('添加菜单项')}
                </Button>
                <Button
                  type='primary'
                  theme='solid'
                  icon={<IconSaveStroked />}
                  onClick={onSubmit}
                >
                  {t('保存自定义菜单')}
                </Button>
                <Input
                  prefix={<IconSearch />}
                  placeholder={t('搜索菜单名称')}
                  value={searchText}
                  onChange={(value) => setSearchText(value)}
                  style={{ width: 250 }}
                  showClear
                />
              </Space>

              <Table
                columns={columns}
                dataSource={filtered}
                rowKey='id'
                pagination={{
                  pageSize: 10,
                  showSizeChanger: false,
                  showQuickJumper: true,
                  showTotal: (total, range) =>
                    t('共 {{total}} 项，当前显示 {{start}}-{{end}} 项', {
                      total,
                      start: range[0],
                      end: range[1],
                    }),
                }}
              />
            </div>
          ) : (
            <Form
              values={inputs}
              getFormApi={(formAPI) => (refForm.current = formAPI)}
            >
              <Form.TextArea
                label={t('自定义菜单配置')}
                extraText={''}
                placeholder={t('为一个 JSON 数组')}
                field={'CustomMenu'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => verifyJSON(value),
                    message: t('不是合法的 JSON 字符串'),
                  },
                ]}
                onChange={(value) =>
                  setInputs({ ...inputs, CustomMenu: value })
                }
              />
            </Form>
          )}
        </Form.Section>

        {editMode === 'json' && (
          <Space>
            <Button
              type='primary'
              icon={<IconSaveStroked />}
              onClick={onSubmit}
            >
              {t('保存自定义菜单')}
            </Button>
          </Space>
        )}
      </Space>

      <Modal
        title={isEdit ? t('编辑菜单项') : t('添加菜单项')}
        visible={modalVisible}
        onOk={handleModalOk}
        onCancel={handleModalCancel}
        width={600}
      >
        <Form getFormApi={(api) => (modalFormRef.current = api)}>
          <Form.Input
            field='name'
            label={t('菜单名称')}
            placeholder={t('显示在侧边栏的文字')}
            rules={[
              { required: true, message: t('请输入菜单名称') },
              { min: 1, message: t('名称不能为空') },
            ]}
          />
          <Form.Input
            field='url'
            label={t('URL链接')}
            placeholder={t(
              '支持 http(s):// 或自定义 schema，可使用 {address} {key} 占位符',
            )}
            rules={[{ required: true, message: t('请输入URL链接') }]}
          />
          <Form.Input
            field='icon'
            label={t('图标')}
            placeholder={t(
              'lucide-react 图标名，例如 Bot、Sparkles、Wand2、BookOpen',
            )}
          />
          <Banner
            type='info'
            description={t(
              '图标名可在 lucide.dev 查询；常用示例：Bot、MessageCircle、Sparkles、Wand2、BookOpen、Compass、Star、Heart、Globe、Send。',
            )}
            style={{ marginTop: 16 }}
          />
        </Form>
      </Modal>
    </Spin>
  );
}
