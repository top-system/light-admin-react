/**
 * <DictTag code="..." value="..."> — renders the dict item's label as a
 * colored Tag. Unknown values fall back to showing the raw value.
 */
import { Tag } from 'antd';
import React, { useEffect, useState } from 'react';
import type { DictItemOption } from '@/types/light-admin/domain';
import {
  loadDictOptions,
  peekDictOptions,
  subscribeDict,
} from '@/utils/dict/cache';

export type DictTagProps = {
  code: string;
  value: string | number | null | undefined;
};

export const DictTag: React.FC<DictTagProps> = ({ code, value }) => {
  const [items, setItems] = useState<DictItemOption[] | undefined>(() =>
    peekDictOptions(code),
  );

  useEffect(() => {
    let alive = true;
    loadDictOptions(code).then((v) => alive && setItems(v));
    const off = subscribeDict(code, () => {
      loadDictOptions(code, true).then((v) => alive && setItems(v));
    });
    return () => {
      alive = false;
      off();
    };
  }, [code]);

  if (value === null || value === undefined || value === '')
    return <span>-</span>;
  const found = items?.find((item) => String(item.value) === String(value));
  if (!found) return <span>{String(value)}</span>;
  return <Tag color={found.tagType || undefined}>{found.label}</Tag>;
};

export default DictTag;
