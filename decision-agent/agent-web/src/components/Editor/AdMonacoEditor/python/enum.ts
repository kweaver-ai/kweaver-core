export const pythonLanguageConfig = {
  keywords: [
    'False',
    'None',
    'True',
    'and',
    'as',
    'assert',
    'async',
    'await',
    'break',
    'class',
    'continue',
    'def',
    'del',
    'elif',
    'else',
    'except',
    'finally',
    'for',
    'from',
    'global',
    'if',
    'import',
    'in',
    'is',
    'lambda',
    'nonlocal',
    'not',
    'or',
    'pass',
    'raise',
    'return',
    'try',
    'while',
    'with',
    'yield'
  ],
  builtins: [
    'abs',
    'all',
    'any',
    'bin',
    'bool',
    'bytearray',
    'bytes',
    'chr',
    'classmethod',
    'complex',
    'delattr',
    'dict',
    'dir',
    'divmod',
    'enumerate',
    'eval',
    'filter',
    'float',
    'format',
    'frozenset',
    'getattr',
    'globals',
    'hasattr',
    'hash',
    'help',
    'hex',
    'id',
    'input',
    'int',
    'isinstance',
    'issubclass',
    'iter',
    'len',
    'list',
    'locals',
    'map',
    'max',
    'memoryview',
    'min',
    'next',
    'object',
    'oct',
    'open',
    'ord',
    'pow',
    'print',
    'property',
    'range',
    'repr',
    'reversed',
    'round',
    'set',
    'setattr',
    'slice',
    'sorted',
    'staticmethod',
    'str',
    'sum',
    'super',
    'tuple',
    'type',
    'vars',
    'zip'
  ],
  operators: [
    '+',
    '-',
    '*',
    '**',
    '/',
    '//',
    '%',
    '@',
    '<<',
    '>>',
    '&',
    '|',
    '^',
    '~',
    '<',
    '>',
    '<=',
    '>=',
    '==',
    '!='
  ],
  symbols: /[=><!~?:&|+\-*\/\^%]+/,
  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,
  tokenizer: {
    root: [
      // 标识符和关键字
      [
        /[a-z_$][\w$]*/,
        {
          cases: {
            '@keywords': 'keyword',
            '@builtins': 'type.identifier',
            '@default': 'identifier'
          }
        }
      ],

      // 空格
      { include: '@whitespace' },

      // 装饰器
      [/@[a-zA-Z_]\w*/, 'decorator'],

      // 数字
      [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
      [/0[xX][0-9a-fA-F]+/, 'number.hex'],
      [/0[oO][0-7]+/, 'number.octal'],
      [/0[bB][0-1]+/, 'number.binary'],
      [/\d+/, 'number'],

      // 分隔符
      [/[{}()\[\]]/, '@brackets'],
      [/[<>](?!@symbols)/, '@brackets'],
      [
        /@symbols/,
        {
          cases: {
            '@operators': 'operator',
            '@default': ''
          }
        }
      ],

      // 字符串
      [/'([^'\\]|\\.)*$/, 'string.invalid'],
      [/"([^"\\]|\\.)*$/, 'string.invalid'],
      [/'/, 'string', '@string_single'],
      [/"/, 'string', '@string_double'],
      [/[fF]'/, 'string', '@string_single'],
      [/[fF]"/, 'string', '@string_double'],
      [/[rR]'/, 'string.regex', '@string_single'],
      [/[rR]"/, 'string.regex', '@string_double']
    ],

    whitespace: [
      [/[ \\t\\r\\n]+/, 'white'],
      [/#.*$/, 'comment']
    ],

    string_single: [
      [/[^\\']+/, 'string'],
      [/@escapes/, 'string.escape'],
      [/\\./, 'string.escape.invalid'],
      [/'/, 'string', '@pop']
    ],

    string_double: [
      [/[^\\"]+/, 'string'],
      [/@escapes/, 'string.escape'],
      [/\\./, 'string.escape.invalid'],
      [/"/, 'string', '@pop']
    ]
  }
};
