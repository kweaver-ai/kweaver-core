/*
  ATTENTION! This file (but not the functions within) is deprecated.

  You should probably add a new file to `./helpers/` instead of adding a new
  function here.

  One-function-per-file is a better pattern than what we have here.

  If you're refactoring something in here, feel free to break it out to a file
  in `./helpers` if you have the time.
*/

import Im, { fromJS, Set } from 'immutable';
import { sanitizeUrl as braintreeSanitizeUrl } from '@braintree/sanitize-url';
import camelCase from 'lodash/camelCase';
import upperFirst from 'lodash/upperFirst';
import _memoize from 'lodash/memoize';
import find from 'lodash/find';
import some from 'lodash/some';
import eq from 'lodash/eq';
import isFunction from 'lodash/isFunction';
import { memoizedSampleFromSchema, memoizedCreateXMLExample } from '../plugins/samples/fn';
import win from '../window';
import cssEscape from 'css.escape';
import getParameterSchema from '../../helpers/get-parameter-schema';
import randomBytes from 'randombytes';
import shaJs from 'sha.js';
import YAML, { JSON_SCHEMA } from 'js-yaml';

const DEFAULT_RESPONSE_KEY = 'default';

export const isImmutable = maybe => Im.Iterable.isIterable(maybe);

export function objectify(thing) {
  if (!isObject(thing)) return {};
  if (isImmutable(thing)) return thing.toJS();
  return thing;
}

export function arrayify(thing) {
  if (!thing) return [];

  if (thing.toArray) return thing.toArray();

  return normalizeArray(thing);
}

export function fromJSOrdered(js) {
  if (isImmutable(js)) {
    return js; // Can't do much here
  }
  if (js instanceof win.File) {
    return js;
  }
  if (!isObject(js)) {
    return js;
  }
  if (Array.isArray(js)) {
    return Im.Seq(js).map(fromJSOrdered).toList();
  }
  if (isFunction(js.entries)) {
    // handle multipart/form-data
    const objWithHashedKeys = createObjWithHashedKeys(js);
    return Im.OrderedMap(objWithHashedKeys).map(fromJSOrdered);
  }
  return Im.OrderedMap(js).map(fromJSOrdered);
}

/**
 * Convert a FormData object into plain object
 * Append a hashIdx and counter to the key name, if multiple exists
 * if single, key name = <original>
 * if multiple, key name = <original><hashIdx><count>
 * @example <caption>single entry for vegetable</caption>
 * fdObj.entries.vegtables: "carrot"
 * // returns newObj.vegetables : "carrot"
 * @example <caption>multiple entries for fruits[]</caption>
 * fdObj.entries.fruits[]: "apple"
 * // returns newObj.fruits[]_**[]1 : "apple"
 * fdObj.entries.fruits[]: "banana"
 * // returns newObj.fruits[]_**[]2 : "banana"
 * fdObj.entries.fruits[]: "grape"
 * // returns newObj.fruits[]_**[]3 : "grape"
 * @param {FormData} fdObj - a FormData object
 * @return {Object} - a plain object
 */
export function createObjWithHashedKeys(fdObj) {
  if (!isFunction(fdObj.entries)) {
    return fdObj; // not a FormData object with iterable
  }
  const newObj = {};
  const hashIdx = '_**[]'; // our internal identifier
  const trackKeys = {};
  for (const pair of fdObj.entries()) {
    if (!newObj[pair[0]] && !(trackKeys[pair[0]] && trackKeys[pair[0]].containsMultiple)) {
      newObj[pair[0]] = pair[1]; // first key name: no hash required
    } else {
      if (!trackKeys[pair[0]]) {
        // initiate tracking key for multiple
        trackKeys[pair[0]] = {
          containsMultiple: true,
          length: 1,
        };
        // "reassign" first pair to matching hashed format for multiple
        const hashedKeyFirst = `${pair[0]}${hashIdx}${trackKeys[pair[0]].length}`;
        newObj[hashedKeyFirst] = newObj[pair[0]];
        // remove non-hashed key of multiple
        delete newObj[pair[0]]; // first
      }
      trackKeys[pair[0]].length += 1;
      const hashedKeyCurrent = `${pair[0]}${hashIdx}${trackKeys[pair[0]].length}`;
      newObj[hashedKeyCurrent] = pair[1];
    }
  }
  return newObj;
}

export function bindToState(obj, state) {
  const newObj = {};
  Object.keys(obj)
    .filter(key => typeof obj[key] === 'function')
    .forEach(key => (newObj[key] = obj[key].bind(null, state)));
  return newObj;
}

export function normalizeArray(arr) {
  if (Array.isArray(arr)) return arr;
  return [arr];
}

export function isFn(fn) {
  return typeof fn === 'function';
}

function isObject(obj) {
  return !!obj && typeof obj === 'object';
}

export function isFunc(thing) {
  return typeof thing === 'function';
}

export function isArray(thing) {
  return Array.isArray(thing);
}

// I've changed memoize libs more than once, so I'm using this a way to make that simpler
export const memoize = _memoize;

export function objMap(obj, fn) {
  return Object.keys(obj).reduce((newObj, key) => {
    newObj[key] = fn(obj[key], key);
    return newObj;
  }, {});
}

export function objReduce(obj, fn) {
  return Object.keys(obj).reduce((newObj, key) => {
    const res = fn(obj[key], key);
    if (res && typeof res === 'object') Object.assign(newObj, res);
    return newObj;
  }, {});
}

export function systemThunkMiddleware(getSystem) {
  return ({ dispatch, getState }) => {
    return next => action => {
      if (typeof action === 'function') {
        return action(getSystem());
      }

      return next(action);
    };
  };
}

export function defaultStatusCode(responses) {
  const codes = responses.keySeq();
  return codes.contains(DEFAULT_RESPONSE_KEY)
    ? DEFAULT_RESPONSE_KEY
    : codes
        .filter(key => (key + '')[0] === '2')
        .sort()
        .first();
}

/**
 * Returns an Immutable List, safely
 * @param {Immutable.Iterable} iterable the iterable to get the key from
 * @param {String|[String]} key either an array of keys, or a single key
 * @returns {Immutable.List} either iterable.get(keys) or an empty Immutable.List
 */
export function getList(iterable, keys) {
  if (!Im.Iterable.isIterable(iterable)) {
    return Im.List();
  }
  const val = iterable.getIn(Array.isArray(keys) ? keys : [keys]);
  return Im.List.isList(val) ? val : Im.List();
}

/**
 * Take an immutable map, and convert to a list.
 * Where the keys are merged with the value objects
 * @param {Immutable.Map} map, the map to convert
 * @param {String} key the key to use, when merging the `key`
 * @returns {Immutable.List}
 */
export function mapToList(map, keyNames = 'key', collectedKeys = Im.Map()) {
  if (!Im.Map.isMap(map) || !map.size) {
    return Im.List();
  }

  if (!Array.isArray(keyNames)) {
    keyNames = [keyNames];
  }

  if (keyNames.length < 1) {
    return map.merge(collectedKeys);
  }

  // I need to avoid `flatMap` from merging in the Maps, as well as the lists
  let list = Im.List();
  const keyName = keyNames[0];
  for (const entry of map.entries()) {
    const [key, val] = entry;
    const nextList = mapToList(val, keyNames.slice(1), collectedKeys.set(keyName, key));
    if (Im.List.isList(nextList)) {
      list = list.concat(nextList);
    } else {
      list = list.push(nextList);
    }
  }

  return list;
}

export function extractFileNameFromContentDispositionHeader(value) {
  const patterns = [
    /filename\*=[^']+'\w*'"([^"]+)";?/i,
    /filename\*=[^']+'\w*'([^;]+);?/i,
    /filename="([^;]*);?"/i,
    /filename=([^;]*);?/i,
  ];

  let responseFilename;
  patterns.some(regex => {
    responseFilename = regex.exec(value);
    return responseFilename !== null;
  });

  if (responseFilename !== null && responseFilename.length > 1) {
    try {
      return decodeURIComponent(responseFilename[1]);
    } catch (e) {
      console.error(e);
    }
  }

  return null;
}

// PascalCase, aka UpperCamelCase
export function pascalCase(str) {
  return upperFirst(camelCase(str));
}

// Remove the ext of a filename, and pascalCase it
export function pascalCaseFilename(filename) {
  return pascalCase(filename.replace(/\.[^./]*$/, ''));
}

// Check if ...
// - new props
// - If immutable, use .is()
// - if in explicit objectList, then compare using _.eq
// - else use ===
export const propChecker = (props, nextProps, objectList = [], ignoreList = []) => {
  if (Object.keys(props).length !== Object.keys(nextProps).length) {
    return true;
  }

  return (
    some(props, (a, name) => {
      if (ignoreList.includes(name)) {
        return false;
      }
      const b = nextProps[name];

      if (Im.Iterable.isIterable(a)) {
        return !Im.is(a, b);
      }

      // Not going to compare objects
      if (typeof a === 'object' && typeof b === 'object') {
        return false;
      }

      return a !== b;
    }) || objectList.some(objectPropName => !eq(props[objectPropName], nextProps[objectPropName]))
  );
};

export const validateMaximum = (val, max) => {
  if (val > max) {
    return `Value must be less than ${max}`;
  }
};

export const validateMinimum = (val, min) => {
  if (val < min) {
    return `Value must be greater than ${min}`;
  }
};

export const validateNumber = val => {
  if (!/^-?\d+(\.?\d+)?$/.test(val)) {
    return 'Value must be a number';
  }
};

export const validateInteger = val => {
  if (!/^-?\d+$/.test(val)) {
    return 'Value must be an integer';
  }
};

export const validateFile = val => {
  if (val && !(val instanceof win.File)) {
    return 'Value must be a file';
  }
};

export const validateBoolean = val => {
  if (!(val === 'true' || val === 'false' || val === true || val === false)) {
    return 'Value must be a boolean';
  }
};

export const validateString = val => {
  if (val && typeof val !== 'string') {
    return 'Value must be a string';
  }
};

export const validateDateTime = val => {
  if (isNaN(Date.parse(val))) {
    return 'Value must be a DateTime';
  }
};

export const validateGuid = val => {
  val = val.toString().toLowerCase();
  if (!/^[{(]?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}[)}]?$/.test(val)) {
    return 'Value must be a Guid';
  }
};

export const validateMaxLength = (val, max) => {
  if (val.length > max) {
    return `Value must be no longer than ${max} character${max !== 1 ? 's' : ''}`;
  }
};

export const validateUniqueItems = (val, uniqueItems) => {
  if (!val) {
    return;
  }
  if (uniqueItems === 'true' || uniqueItems === true) {
    const list = fromJS(val);
    const set = list.toSet();
    const hasDuplicates = val.length > set.size;
    if (hasDuplicates) {
      let errorsPerIndex = Set();
      list.forEach((item, i) => {
        if (list.filter(v => (isFunc(v.equals) ? v.equals(item) : v === item)).size > 1) {
          errorsPerIndex = errorsPerIndex.add(i);
        }
      });
      if (errorsPerIndex.size !== 0) {
        return errorsPerIndex.map(i => ({ index: i, error: 'No duplicates allowed.' })).toArray();
      }
    }
  }
};

export const validateMinItems = (val, min) => {
  if ((!val && min >= 1) || (val && val.length < min)) {
    return `Array must contain at least ${min} item${min === 1 ? '' : 's'}`;
  }
};

export const validateMaxItems = (val, max) => {
  if (val && val.length > max) {
    return `Array must not contain more then ${max} item${max === 1 ? '' : 's'}`;
  }
};

export const validateMinLength = (val, min) => {
  if (val.length < min) {
    return `Value must be at least ${min} character${min !== 1 ? 's' : ''}`;
  }
};

export const validatePattern = (val, rxPattern) => {
  const patt = new RegExp(rxPattern);
  if (!patt.test(val)) {
    return 'Value must follow pattern ' + rxPattern;
  }
};

function validateValueBySchema(value, schema, requiredByParam, bypassRequiredCheck, parameterContentMediaType) {
  if (!schema) return [];
  const errors = [];
  const nullable = schema.get('nullable');
  const requiredBySchema = schema.get('required');
  const maximum = schema.get('maximum');
  const minimum = schema.get('minimum');
  const type = schema.get('type');
  const format = schema.get('format');
  const maxLength = schema.get('maxLength');
  const minLength = schema.get('minLength');
  const uniqueItems = schema.get('uniqueItems');
  const maxItems = schema.get('maxItems');
  const minItems = schema.get('minItems');
  const pattern = schema.get('pattern');

  const schemaRequiresValue = requiredByParam || requiredBySchema === true;
  const hasValue = value !== undefined && value !== null;
  const isValidEmpty = !schemaRequiresValue && !hasValue;

  const needsExplicitConstraintValidation = hasValue && type === 'array';

  const requiresFurtherValidation = schemaRequiresValue || needsExplicitConstraintValidation || !isValidEmpty;

  const isValidNullable = nullable && value === null;

  // will not be included in the request or [schema / value] does not [allow / require] further analysis.
  const noFurtherValidationNeeded = isValidNullable || !type || !requiresFurtherValidation;

  if (noFurtherValidationNeeded) {
    return [];
  }

  // Further this point the parameter is considered worth to validate
  const stringCheck = type === 'string' && value;
  const arrayCheck = type === 'array' && Array.isArray(value) && value.length;
  const arrayListCheck = type === 'array' && Im.List.isList(value) && value.count();
  const arrayStringCheck = type === 'array' && typeof value === 'string' && value;
  const fileCheck = type === 'file' && value instanceof win.File;
  const booleanCheck = type === 'boolean' && (value || value === false);
  const numberCheck = type === 'number' && (value || value === 0);
  const integerCheck = type === 'integer' && (value || value === 0);
  const objectCheck = type === 'object' && typeof value === 'object' && value !== null;
  const objectStringCheck = type === 'object' && typeof value === 'string' && value;

  const allChecks = [
    stringCheck,
    arrayCheck,
    arrayListCheck,
    arrayStringCheck,
    fileCheck,
    booleanCheck,
    numberCheck,
    integerCheck,
    objectCheck,
    objectStringCheck,
  ];

  const passedAnyCheck = allChecks.some(v => !!v);

  if (schemaRequiresValue && !passedAnyCheck && !bypassRequiredCheck) {
    errors.push('Required field is not provided');
    return errors;
  }
  if (type === 'object' && (parameterContentMediaType === null || parameterContentMediaType === 'application/json')) {
    let objectVal = value;
    if (typeof value === 'string') {
      try {
        objectVal = JSON.parse(value);
      } catch (e) {
        errors.push('Parameter string value must be valid JSON');
        return errors;
      }
    }
    if (schema && schema.has('required') && isFunc(requiredBySchema.isList) && requiredBySchema.isList()) {
      requiredBySchema.forEach(key => {
        if (objectVal[key] === undefined) {
          errors.push({ propKey: key, error: 'Required property not found' });
        }
      });
    }
    if (schema && schema.has('properties')) {
      schema.get('properties').forEach((val, key) => {
        const errs = validateValueBySchema(objectVal[key], val, false, bypassRequiredCheck, parameterContentMediaType);
        errors.push(...errs.map(error => ({ propKey: key, error })));
      });
    }
  }

  if (pattern) {
    const err = validatePattern(value, pattern);
    if (err) errors.push(err);
  }

  if (minItems) {
    if (type === 'array') {
      const err = validateMinItems(value, minItems);
      if (err) errors.push(err);
    }
  }

  if (maxItems) {
    if (type === 'array') {
      const err = validateMaxItems(value, maxItems);
      if (err) errors.push({ needRemove: true, error: err });
    }
  }

  if (uniqueItems) {
    if (type === 'array') {
      const errorPerItem = validateUniqueItems(value, uniqueItems);
      if (errorPerItem) errors.push(...errorPerItem);
    }
  }

  if (maxLength || maxLength === 0) {
    const err = validateMaxLength(value, maxLength);
    if (err) errors.push(err);
  }

  if (minLength) {
    const err = validateMinLength(value, minLength);
    if (err) errors.push(err);
  }

  if (maximum || maximum === 0) {
    const err = validateMaximum(value, maximum);
    if (err) errors.push(err);
  }

  if (minimum || minimum === 0) {
    const err = validateMinimum(value, minimum);
    if (err) errors.push(err);
  }

  if (type === 'string') {
    let err;
    if (format === 'date-time') {
      err = validateDateTime(value);
    } else if (format === 'uuid') {
      err = validateGuid(value);
    } else {
      err = validateString(value);
    }
    if (!err) return errors;
    errors.push(err);
  } else if (type === 'boolean') {
    const err = validateBoolean(value);
    if (!err) return errors;
    errors.push(err);
  } else if (type === 'number') {
    const err = validateNumber(value);
    if (!err) return errors;
    errors.push(err);
  } else if (type === 'integer') {
    const err = validateInteger(value);
    if (!err) return errors;
    errors.push(err);
  } else if (type === 'array') {
    if (!(arrayCheck || arrayListCheck)) {
      return errors;
    }
    if (value) {
      value.forEach((item, i) => {
        const errs = validateValueBySchema(
          item,
          schema.get('items'),
          false,
          bypassRequiredCheck,
          parameterContentMediaType
        );
        errors.push(...errs.map(err => ({ index: i, error: err })));
      });
    }
  } else if (type === 'file') {
    const err = validateFile(value);
    if (!err) return errors;
    errors.push(err);
  }

  return errors;
}

// validation of parameters before execute
export const validateParam = (param, value, { isOAS3 = false, bypassRequiredCheck = false } = {}) => {
  const paramRequired = param.get('required');

  const { schema: paramDetails, parameterContentMediaType } = getParameterSchema(param, { isOAS3 });

  return validateValueBySchema(value, paramDetails, paramRequired, bypassRequiredCheck, parameterContentMediaType);
};

const getXmlSampleSchema = (schema, config, exampleOverride) => {
  if (schema && !schema.xml) {
    schema.xml = {};
  }
  if (schema && !schema.xml.name) {
    if (!schema.$$ref && (schema.type || schema.items || schema.properties || schema.additionalProperties)) {
      return '<?xml version="1.0" encoding="UTF-8"?>\n<!-- XML example cannot be generated; root element name is undefined -->';
    }
    if (schema.$$ref) {
      const match = schema.$$ref.match(/\S*\/(\S+)$/);
      schema.xml.name = match[1];
    }
  }

  return memoizedCreateXMLExample(schema, config, exampleOverride);
};

const shouldStringifyTypesConfig = [
  {
    when: /json/,
    shouldStringifyTypes: ['string'],
  },
];

const defaultStringifyTypes = ['object'];

const getStringifiedSampleForSchema = (schema, config, contentType, exampleOverride) => {
  // console.log('schema, config, contentType, exampleOverride', schema, config, contentType, exampleOverride);
  const res = memoizedSampleFromSchema(schema, config, exampleOverride);
  const resType = typeof res;

  const typesToStringify = shouldStringifyTypesConfig.reduce(
    (types, nextConfig) => (nextConfig.when.test(contentType) ? [...types, ...nextConfig.shouldStringifyTypes] : types),
    defaultStringifyTypes
  );

  return some(typesToStringify, x => x === resType) ? JSON.stringify(res, null, 2) : res;
};

const getYamlSampleSchema = (schema, config, contentType, exampleOverride) => {
  const jsonExample = getStringifiedSampleForSchema(schema, config, contentType, exampleOverride);
  let yamlString;
  try {
    yamlString = YAML.dump(
      YAML.load(jsonExample),
      {
        lineWidth: -1, // don't generate line folds
      },
      { schema: JSON_SCHEMA }
    );
    if (yamlString[yamlString.length - 1] === '\n') {
      yamlString = yamlString.slice(0, yamlString.length - 1);
    }
  } catch (e) {
    console.error(e);
    return 'error: could not generate yaml example';
  }
  return yamlString.replace(/\t/g, '  ');
};

export const getSampleSchema = (schema, contentType = '', config = {}, exampleOverride = undefined) => {
  if (schema && isFunc(schema.toJS)) schema = schema.toJS();

  if (exampleOverride && isFunc(exampleOverride.toJS)) exampleOverride = exampleOverride.toJS();

  if (/xml/.test(contentType)) {
    return getXmlSampleSchema(schema, config, exampleOverride);
  }
  if (/(yaml|yml)/.test(contentType)) {
    return getYamlSampleSchema(schema, config, contentType, exampleOverride);
  }
  return getStringifiedSampleForSchema(schema, config, contentType, exampleOverride);
};

export const parseSearch = () => {
  const map = {};
  const search = win.location.search;

  if (!search) return {};

  if (search !== '') {
    const params = search.substr(1).split('&');

    for (let i in params) {
      if (!Object.prototype.hasOwnProperty.call(params, i)) {
        continue;
      }
      i = params[i].split('=');
      map[decodeURIComponent(i[0])] = (i[1] && decodeURIComponent(i[1])) || '';
    }
  }

  return map;
};

export const serializeSearch = searchMap => {
  return Object.keys(searchMap)
    .map(k => {
      return encodeURIComponent(k) + '=' + encodeURIComponent(searchMap[k]);
    })
    .join('&');
};

export const btoa = str => {
  let buffer;

  if (str instanceof Buffer) {
    buffer = str;
  } else {
    buffer = Buffer.from(str.toString(), 'utf-8');
  }

  return buffer.toString('base64');
};

export const sorters = {
  operationsSorter: {
    alpha: (a, b) => a.get('path').localeCompare(b.get('path')),
    method: (a, b) => a.get('method').localeCompare(b.get('method')),
  },
  tagsSorter: {
    alpha: (a, b) => a.localeCompare(b),
  },
};

export const buildFormData = data => {
  const formArr = [];

  for (const name in data) {
    const val = data[name];
    if (val !== undefined && val !== '') {
      formArr.push([name, '=', encodeURIComponent(val).replace(/%20/g, '+')].join(''));
    }
  }
  return formArr.join('&');
};

export const shallowEqualKeys = (a, b, keys) => {
  return !!find(keys, key => {
    return eq(a[key], b[key]);
  });
};

export function sanitizeUrl(url) {
  if (typeof url !== 'string' || url === '') {
    return '';
  }

  return braintreeSanitizeUrl(url);
}

export function requiresValidationURL(uri) {
  if (!uri || uri.indexOf('localhost') >= 0 || uri.indexOf('127.0.0.1') >= 0 || uri === 'none') {
    return false;
  }
  return true;
}

export function getAcceptControllingResponse(responses) {
  if (!Im.OrderedMap.isOrderedMap(responses)) {
    // wrong type!
    return null;
  }

  if (!responses.size) {
    // responses is empty
    return null;
  }

  const suitable2xxResponse = responses.find((res, k) => {
    return k.startsWith('2') && Object.keys(res.get('content') || {}).length > 0;
  });

  // try to find a suitable `default` responses
  const defaultResponse = responses.get('default') || Im.OrderedMap();
  const defaultResponseMediaTypes = (defaultResponse.get('content') || Im.OrderedMap()).keySeq().toJS();
  const suitableDefaultResponse = defaultResponseMediaTypes.length ? defaultResponse : null;

  return suitable2xxResponse || suitableDefaultResponse;
}

// suitable for use in URL fragments
export const createDeepLinkPath = str =>
  typeof str === 'string' || str instanceof String ? str.trim().replace(/\s/g, '%20') : '';
// suitable for use in CSS classes and ids
export const escapeDeepLinkPath = str => cssEscape(createDeepLinkPath(str).replace(/%20/g, '_'));

export const getExtensions = defObj => defObj.filter((v, k) => /^x-/.test(k));
export const getCommonExtensions = defObj =>
  defObj.filter((v, k) => /^pattern|maxLength|minLength|maximum|minimum/.test(k));

// Deeply strips a specific key from an object.
//
// `predicate` can be used to discriminate the stripping further,
// by preserving the key's place in the object based on its value.
export function deeplyStripKey(input, keyToStrip, predicate = () => true) {
  if (typeof input !== 'object' || Array.isArray(input) || input === null || !keyToStrip) {
    return input;
  }

  const obj = Object.assign({}, input);

  Object.keys(obj).forEach(k => {
    if (k === keyToStrip && predicate(obj[k], k)) {
      delete obj[k];
      return;
    }
    obj[k] = deeplyStripKey(obj[k], keyToStrip, predicate);
  });

  return obj;
}

export function stringify(thing: any) {
  if (typeof thing === 'string') {
    return thing;
  }

  if (thing && thing.toJS) {
    thing = thing.toJS();
  }

  if (typeof thing === 'object' && thing !== null) {
    try {
      return JSON.stringify(thing, null, 2);
    } catch (e) {
      return String(thing);
    }
  }

  if (thing === null || thing === undefined) {
    return '';
  }

  return thing.toString();
}

export function numberToString(thing) {
  if (typeof thing === 'number') {
    return thing.toString();
  }

  return thing;
}

export function paramToIdentifier(param, { returnAll = false, allowHashes = true } = {}) {
  if (!Im.Map.isMap(param)) {
    throw new Error('paramToIdentifier: received a non-Im.Map parameter as input');
  }
  const paramName = param.get('name');
  const paramIn = param.get('in');

  const generatedIdentifiers = [];

  // Generate identifiers in order of most to least specificity

  if (param && param.hashCode && paramIn && paramName && allowHashes) {
    generatedIdentifiers.push(`${paramIn}.${paramName}.hash-${param.hashCode()}`);
  }

  if (paramIn && paramName) {
    generatedIdentifiers.push(`${paramIn}.${paramName}`);
  }

  generatedIdentifiers.push(paramName);

  // Return the most preferred identifier, or all if requested

  return returnAll ? generatedIdentifiers : generatedIdentifiers[0] || '';
}

export function paramToValue(param, paramValues) {
  const allIdentifiers = paramToIdentifier(param, { returnAll: true });

  // Map identifiers to values in the provided value hash, filter undefined values,
  // and return the first value found
  const values = allIdentifiers
    .map(id => {
      return paramValues[id];
    })
    .filter(value => value !== undefined);

  return values[0];
}

// adapted from https://auth0.com/docs/flows/guides/auth-code-pkce/includes/create-code-verifier
export function generateCodeVerifier() {
  return b64toB64UrlEncoded(randomBytes(32).toString('base64'));
}

export function createCodeChallenge(codeVerifier) {
  return b64toB64UrlEncoded(shaJs('sha256').update(codeVerifier).digest('base64'));
}

function b64toB64UrlEncoded(str) {
  return str.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

export const isEmptyValue = value => {
  if (!value) {
    return true;
  }

  if (isImmutable(value) && value.isEmpty()) {
    return true;
  }

  return false;
};
