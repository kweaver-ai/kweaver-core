import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const currentFile = fileURLToPath(import.meta.url);
const currentDir = path.dirname(currentFile);
const outputFile = 'aggregate.md';

const sourceFiles = [
  'index.md',
  'install-local-sdk-package.md',
];

function readMarkdown(file) {
  return fs.readFileSync(path.join(currentDir, file), 'utf8').replace(/\r\n/g, '\n').trimEnd();
}

function firstHeading(markdown, fallback) {
  const match = markdown.match(/^#\s+(.+)$/m);
  return match?.[1]?.trim() ?? fallback;
}

function githubLikeSlug(text, usedSlugs) {
  const base = text
    .trim()
    .toLowerCase()
    .replace(/[`'"“”‘’]/g, '')
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
  const normalized = base || 'section';
  const seen = usedSlugs.get(normalized) ?? 0;
  usedSlugs.set(normalized, seen + 1);
  return seen === 0 ? normalized : `${normalized}-${seen}`;
}

function shiftHeadings(markdown) {
  let inFence = false;
  return markdown
    .split('\n')
    .map((line) => {
      if (/^\s*(```|~~~)/.test(line)) {
        inFence = !inFence;
        return line;
      }
      if (!inFence && /^#{1,5}\s+/.test(line)) {
        return `#${line}`;
      }
      return line;
    })
    .join('\n');
}

function rewriteLinks(markdown, anchors) {
  return markdown.replace(/\]\((\.\/)?([^)#]+\.md)(#[^)]+)?\)/g, (match, prefix, file, hash) => {
    if (hash) return match;
    const anchor = anchors.get(file);
    if (!anchor) return match;
    return `](#${anchor})`;
  });
}

const usedSlugs = new Map();
const sources = sourceFiles.map((file) => {
  const markdown = readMarkdown(file);
  const title = firstHeading(markdown, file);
  const anchor = githubLikeSlug(title, usedSlugs);
  return { file, title, anchor, markdown };
});

const anchors = new Map(sources.map((source) => [source.file, source.anchor]));
const output = [
  '<!-- 请勿直接编辑：本文件由 docs/user_manual/setup 下的 `make aggregate` 生成。 -->',
  '<!-- 来源文件：index.md, install-local-sdk-package.md。 -->',
  '',
  '# 安装与部署准备（聚合版）',
  '',
  '> 本文件由脚本生成，请不要直接修改本文件；如需调整内容，请修改分文件文档后运行 `make -C docs/user_manual/setup aggregate`。',
  '',
  '## 目录',
  '',
  ...sources.map((source) => `- [${source.title}](#${source.anchor})`),
  '',
  ...sources.flatMap((source) => {
    const content = shiftHeadings(rewriteLinks(source.markdown, anchors));
    return [
      '',
      `<!-- 来源：${source.file} -->`,
      '',
      content,
      '',
    ];
  }),
].join('\n');

fs.writeFileSync(path.join(currentDir, outputFile), `${output.trimEnd()}\n`);
console.log(`Generated ${path.relative(process.cwd(), path.join(currentDir, outputFile))}`);

