import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { ResizableTable } from './index';

test('dragging a header edge changes the column width', () => {
  // jsdom does not provide PointerEvent; MouseEvent carries clientX for this test.
  Object.defineProperty(window, 'PointerEvent', {
    value: MouseEvent,
    configurable: true,
  });
  render(
    <ResizableTable
      rowKey="id"
      pagination={false}
      dataSource={[{ id: '1', name: 'A' }]}
      columns={[{ title: 'Name', dataIndex: 'name', width: 120 }]}
    />,
  );

  const handle = screen.getByRole('button', { name: '调整列宽' });
  const header = handle.closest('th') as HTMLElement;
  expect(header.style.width).toBe('120px');

  fireEvent.pointerDown(handle, { clientX: 100 });
  fireEvent.pointerMove(window, { clientX: 150 });
  fireEvent.pointerUp(window);

  expect(header.style.width).toBe('170px');
});
