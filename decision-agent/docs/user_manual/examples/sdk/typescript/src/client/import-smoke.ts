import { configureKweaver, kweaver } from '../shared.js';

// 目标：验证用户项目安装 SDK 后，可以通过推荐入口导入 kweaver。
configureKweaver();
console.log(typeof kweaver.getClient);
