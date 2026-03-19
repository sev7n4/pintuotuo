import { useUIStore } from '../uiStore'

// 模拟localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('uiStore', () => {
  beforeEach(() => {
    // 清除localStorage
    localStorage.clear()
    // 重置store状态
    useUIStore.setState({
      theme: 'light',
      sidebarCollapsed: false,
      notifications: [],
    })
  })

  test('初始状态正确', () => {
    const state = useUIStore.getState()
    expect(state.theme).toBe('light')
    expect(state.sidebarCollapsed).toBe(false)
    expect(state.notifications).toEqual([])
  })

  test('从localStorage加载主题', () => {
    // 设置localStorage中的主题
    localStorage.setItem('theme', 'dark')
    
    // 重置store以重新加载
    useUIStore.setState({
      theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
      sidebarCollapsed: false,
      notifications: [],
    })
    
    const state = useUIStore.getState()
    expect(state.theme).toBe('dark')
  })

  test('setTheme 切换主题', () => {
    const store = useUIStore.getState()
    store.setTheme('dark')
    
    const newState = useUIStore.getState()
    expect(newState.theme).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')
    
    store.setTheme('light')
    const lightState = useUIStore.getState()
    expect(lightState.theme).toBe('light')
    expect(localStorage.getItem('theme')).toBe('light')
  })

  test('toggleSidebar 切换侧边栏状态', () => {
    const store = useUIStore.getState()
    
    // 初始状态应该是false
    expect(store.sidebarCollapsed).toBe(false)
    
    // 切换到true
    store.toggleSidebar()
    const collapsedState = useUIStore.getState()
    expect(collapsedState.sidebarCollapsed).toBe(true)
    
    // 切换回false
    store.toggleSidebar()
    const expandedState = useUIStore.getState()
    expect(expandedState.sidebarCollapsed).toBe(false)
  })

  test('addNotification 添加通知', () => {
    const store = useUIStore.getState()
    
    // 添加成功通知
    store.addNotification('success', '操作成功')
    
    const newState = useUIStore.getState()
    expect(newState.notifications).toHaveLength(1)
    expect(newState.notifications[0].type).toBe('success')
    expect(newState.notifications[0].message).toBe('操作成功')
    expect(newState.notifications[0].id).toContain('success-')
    
    // 添加错误通知
    store.addNotification('error', '操作失败')
    const errorState = useUIStore.getState()
    expect(errorState.notifications).toHaveLength(2)
    expect(errorState.notifications[1].type).toBe('error')
    expect(errorState.notifications[1].message).toBe('操作失败')
  })

  test('removeNotification 删除通知', () => {
    const store = useUIStore.getState()
    
    // 添加通知
    store.addNotification('success', '操作成功')
    const notificationId = useUIStore.getState().notifications[0].id
    
    // 删除通知
    store.removeNotification(notificationId)
    
    const newState = useUIStore.getState()
    expect(newState.notifications).toEqual([])
  })

  test('clearNotifications 清除所有通知', () => {
    const store = useUIStore.getState()
    
    // 添加多个通知
    store.addNotification('success', '操作成功1')
    store.addNotification('error', '操作失败1')
    store.addNotification('info', '信息1')
    
    expect(useUIStore.getState().notifications).toHaveLength(3)
    
    // 清除所有通知
    store.clearNotifications()
    
    const newState = useUIStore.getState()
    expect(newState.notifications).toEqual([])
  })

  test('通知自动移除', async () => {
    // 模拟setTimeout
    jest.useFakeTimers()
    
    const store = useUIStore.getState()
    store.addNotification('success', '操作成功')
    
    expect(useUIStore.getState().notifications).toHaveLength(1)
    
    // 快进3秒
    jest.advanceTimersByTime(3000)
    
    // 通知应该被自动移除
    expect(useUIStore.getState().notifications).toEqual([])
    
    // 恢复真实的setTimeout
    jest.useRealTimers()
  })
})
