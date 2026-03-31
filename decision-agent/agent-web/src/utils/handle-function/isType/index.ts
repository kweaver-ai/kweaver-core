/**
 * 判断数据类型
 */

// 通用
const baseBool = obj => Object.prototype.toString.call(obj);

// 字符串
const isString = str => typeof str === 'string';

// 数字
const isNumber = num => typeof num === 'number';

// bigint
const isBigint = num => typeof num === 'bigint';

// 布尔
const isBoolean = bool => typeof bool === 'boolean';

// 函数
const isFunction = func => typeof func === 'function';

// undefined
const isUndefined = obj => typeof obj === 'undefined';

// 数组
const isArray = arr => baseBool(arr) === '[object Array]';

// 严格的对象{}
const isObject = obj => baseBool(obj) === '[object Object]';

// undefined
const isNull = obj => baseBool(obj) === '[object Null]';

// 是否已经定义
const isDef = value => value !== undefined && value !== null;

// 是否是JSON字符串
const isJSONString = str => {
  try {
    if (JSON.parse(str) instanceof Object) {
      return true;
    }
  } catch (e) {
    return false;
  }
};

export {
  baseBool,
  isString,
  isNumber,
  isBigint,
  isBoolean,
  isFunction,
  isArray,
  isObject,
  isUndefined,
  isNull,
  isDef,
  isJSONString,
};
