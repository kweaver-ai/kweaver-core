import { EventEmitter } from 'events';

export enum ComponentIDEnum {
  Global = 'global',
  AgentUsage = 'agentUsage',
  APIDoc = 'apiDoc',
}

// 存储所有组件的 EventEmitter 实例
const componentEmitters = new Map<string, EventEmitter>();

/**
 * 获取或创建组件级别的 EventEmitter 单例
 * @param componentId 组件唯一标识
 * @returns EventEmitter 实例
 */
export function getComponentEventEmitter(componentId: string): EventEmitter {
  if (!componentEmitters.has(componentId)) {
    const emitter = new EventEmitter();
    // 设置最大监听器数量，避免内存泄漏警告
    emitter.setMaxListeners(50);
    componentEmitters.set(componentId, emitter);
  }
  return componentEmitters.get(componentId)!;
}

/**
 * 清理组件的事件监听器
 * @param componentId 组件唯一标识
 */
export function cleanupComponentEmitter(componentId: string): void {
  const emitter = componentEmitters.get(componentId);
  if (emitter) {
    emitter.removeAllListeners();
    componentEmitters.delete(componentId);
  }
}
