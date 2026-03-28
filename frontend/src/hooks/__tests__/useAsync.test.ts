import { renderHook, act, waitFor } from '@testing-library/react';
import { useAsync } from '../useAsync';

describe('useAsync hook', () => {
  test('initial state is correct', () => {
    const asyncFunction = jest.fn();

    const { result } = renderHook(() => useAsync(asyncFunction, false));

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  test('executes async function immediately when immediate is true', async () => {
    const mockData = { id: 1, name: 'Test Data' };
    const asyncFunction = jest.fn().mockResolvedValue(mockData);

    const { result } = renderHook(() => useAsync(asyncFunction, true));

    // 初始状态：loading为true
    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();

    // 等待异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 最终状态：data为mockData，loading为false，error为null
    expect(result.current.data).toEqual(mockData);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(asyncFunction).toHaveBeenCalledTimes(1);
  });

  test('does not execute async function immediately when immediate is false', () => {
    const asyncFunction = jest.fn();

    renderHook(() => useAsync(asyncFunction, false));

    expect(asyncFunction).not.toHaveBeenCalled();
  });

  test('handles async function error', async () => {
    const mockError = new Error('Test Error');
    const asyncFunction = jest.fn().mockRejectedValue(mockError);

    const { result } = renderHook(() => useAsync(asyncFunction, true));

    // 初始状态：loading为true
    expect(result.current.loading).toBe(true);

    // 等待异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 最终状态：error为mockError，loading为false，data为null
    expect(result.current.error).toEqual(mockError);
    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(asyncFunction).toHaveBeenCalledTimes(1);
  });

  test('handles non-Error rejection', async () => {
    const mockError = 'Test Error String';
    const asyncFunction = jest.fn().mockRejectedValue(mockError);

    const { result } = renderHook(() => useAsync(asyncFunction, true));

    // 等待异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 最终状态：error为Error对象，loading为false，data为null
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Unknown error');
    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(asyncFunction).toHaveBeenCalledTimes(1);
  });

  test('refetch method executes async function', async () => {
    const mockData1 = { id: 1, name: 'Test Data 1' };
    const mockData2 = { id: 2, name: 'Test Data 2' };
    const asyncFunction = jest
      .fn()
      .mockResolvedValueOnce(mockData1)
      .mockResolvedValueOnce(mockData2);

    const { result } = renderHook(() => useAsync(asyncFunction, true));

    // 等待第一次异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData1);
    expect(asyncFunction).toHaveBeenCalledTimes(1);

    // 调用refetch方法
    await act(async () => {
      await result.current.refetch();
    });

    // 等待第二次异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 最终状态：data为mockData2，loading为false，error为null
    expect(result.current.data).toEqual(mockData2);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(asyncFunction).toHaveBeenCalledTimes(2);
  });

  test('refetch method handles error', async () => {
    const mockData = { id: 1, name: 'Test Data' };
    const mockError = new Error('Test Error');
    const asyncFunction = jest
      .fn()
      .mockResolvedValueOnce(mockData)
      .mockRejectedValueOnce(mockError);

    const { result } = renderHook(() => useAsync(asyncFunction, true));

    // 等待第一次异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData);
    expect(asyncFunction).toHaveBeenCalledTimes(1);

    // 调用refetch方法
    await act(async () => {
      await result.current.refetch();
    });

    // 等待第二次异步操作完成
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 最终状态：error为mockError，loading为false，data为null
    expect(result.current.error).toEqual(mockError);
    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(asyncFunction).toHaveBeenCalledTimes(2);
  });
});
