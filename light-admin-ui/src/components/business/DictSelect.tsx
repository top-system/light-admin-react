/**
 * <DictSelect code="..."> — Ant Select wired to a dict code. Uses the shared
 * cache so repeated mounts coalesce into one network request, and reacts to
 * `dict-change` ws events (which call `invalidateDict`).
 */
import { Select, type SelectProps } from 'antd';
import React, { useEffect, useState } from 'react';
import type { DictItemOption } from '@/types/light-admin/domain';
import {
  loadDictOptions,
  peekDictOptions,
  subscribeDict,
} from '@/utils/dict/cache';

export type DictSelectProps = Omit<SelectProps<string | number>, 'options'> & {
  code: string;
};

export const DictSelect: React.FC<DictSelectProps> = ({ code, ...rest }) => {
  const [options, setOptions] = useState<DictItemOption[] | undefined>(() =>
    peekDictOptions(code),
  );
  const [loading, setLoading] = useState(!options);

  useEffect(() => {
    let alive = true;
    const load = () => {
      setLoading(true);
      loadDictOptions(code)
        .then((v) => alive && setOptions(v))
        .finally(() => alive && setLoading(false));
    };
    load();
    const off = subscribeDict(code, () => {
      // The cache may have just invalidated — reload.
      loadDictOptions(code, true).then((v) => alive && setOptions(v));
    });
    return () => {
      alive = false;
      off();
    };
  }, [code]);

  return <Select loading={loading} allowClear {...rest} options={options} />;
};

export default DictSelect;
