/**
 * sessionStorage封装, 对存入的数据进行url编码
 */
class SessionStore {
  private store: Storage;

  constructor() {
    this.store = window.sessionStorage;
  }

  // 解码取出
  get(key: string): any {
    const data = this.store.getItem(key);
    if (!data) return null;
    return JSON.parse(data);
  }

  // 编码存入
  set(key: string, data: any): void {
    const encodeData = JSON.stringify(data);
    this.store.setItem(key, encodeData);
  }

  // 删除
  remove(key: string): void {
    this.store.removeItem(key);
  }

  // 清空sessionStorage
  clear(): void {
    this.store.clear();
  }
}

const sessionStore = new SessionStore();

export { sessionStore };
