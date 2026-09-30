import { ProTable, type ProTableProps } from '@ant-design/pro-components';
import { Table, type TableProps } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { createContext, useContext, useMemo, useState } from 'react';

type ResizeContextValue = (key: string, width: number) => void;
const ResizeContext = createContext<ResizeContextValue>(() => {});
const MIN_WIDTH = 80;
const DEFAULT_WIDTH = 160;

type HeaderCellProps = React.ThHTMLAttributes<HTMLTableCellElement> & {
  'data-resize-key'?: string;
  width?: number;
};

function HeaderCell({ children, width, ...props }: HeaderCellProps) {
  const resize = useContext(ResizeContext);
  const key = props['data-resize-key'];
  const onPointerDown = (event: React.PointerEvent<HTMLButtonElement>) => {
    if (!key) return;
    event.preventDefault();
    event.stopPropagation();
    const startX = event.clientX;
    const startWidth = width ?? DEFAULT_WIDTH;
    const move = (moveEvent: PointerEvent) =>
      resize(key, Math.max(MIN_WIDTH, startWidth + moveEvent.clientX - startX));
    const stop = () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', stop);
      window.removeEventListener('pointercancel', stop);
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', stop, { once: true });
    window.addEventListener('pointercancel', stop, { once: true });
  };
  return (
    <th {...props} style={{ ...props.style, position: 'relative', width }}>
      {children}
      {key && (
        <button
          type="button"
          aria-label="调整列宽"
          onPointerDown={onPointerDown}
          onClick={(event) => event.stopPropagation()}
          onKeyDown={(event) => {
            if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
            event.preventDefault();
            event.stopPropagation();
            resize(
              key,
              Math.max(
                MIN_WIDTH,
                (width ?? DEFAULT_WIDTH) +
                  (event.key === 'ArrowRight' ? 10 : -10),
              ),
            );
          }}
          style={{
            position: 'absolute',
            right: 0,
            top: 0,
            bottom: 0,
            width: 8,
            padding: 0,
            border: 0,
            background: 'transparent',
            cursor: 'col-resize',
            touchAction: 'none',
            zIndex: 2,
          }}
        />
      )}
    </th>
  );
}

function useColumns<
  T extends {
    width?: number | string;
    key?: React.Key;
    dataIndex?: unknown;
    onHeaderCell?: (...args: any[]) => any;
  },
>(columns?: readonly T[]) {
  const [widths, setWidths] = useState<Record<string, number>>({});
  const resize = (key: string, width: number) =>
    setWidths((old) => ({ ...old, [key]: width }));
  const resized = useMemo(
    () =>
      columns?.map((column, index) => {
        const key = String(column.key ?? column.dataIndex ?? index);
        const initial =
          typeof column.width === 'number' ? column.width : DEFAULT_WIDTH;
        const width = widths[key] ?? initial;
        const originalHeaderCell = column.onHeaderCell;
        return {
          ...column,
          width,
          onHeaderCell: (...args: any[]): any => ({
            ...originalHeaderCell?.(...args),
            width,
            'data-resize-key': key,
          }),
        };
      }),
    [columns, widths],
  );
  return { resized, resize };
}

export function ResizableTable<T extends object>(props: TableProps<T>) {
  const { resized, resize } = useColumns(props.columns);
  return (
    <ResizeContext.Provider value={resize}>
      <Table<T>
        {...props}
        columns={resized as ColumnsType<T>}
        components={{
          ...props.components,
          header: { ...props.components?.header, cell: HeaderCell },
        }}
      />
    </ResizeContext.Provider>
  );
}

export function ResizableProTable<
  T extends Record<string, any>,
  P extends Record<string, any> = Record<string, any>,
>(props: ProTableProps<T, P>) {
  const { resized, resize } = useColumns(props.columns);
  return (
    <ResizeContext.Provider value={resize}>
      <ProTable<T, P>
        {...props}
        columns={resized as typeof props.columns}
        components={{
          ...props.components,
          header: { ...props.components?.header, cell: HeaderCell },
        }}
      />
    </ResizeContext.Provider>
  );
}
