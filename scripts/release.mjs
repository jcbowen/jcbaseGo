#!/usr/bin/env node
// jcbaseGo 自动发版脚本。
//
// jcbaseGo 是 Go 库项目，发版 = 打 vX.Y.Z 轻量 tag 并（可选）推送 origin，
// 不构建镜像、不导出 tar。流程（VSCode 任务入口见 .vscode/tasks.json）：
//   [1/3] 版本选择：读取本地 v* tag，按 patch/minor/major × 正式/-alpha/-beta/-rc
//         生成候选，方向键菜单两步选择，也可手动输入完整版本号；
//   [2/3] 轻量门禁：gofmt -l . + go build/vet/test ./...（--skip-test 只跳过 test）；
//   [3/3] 打 tag：创建轻量 tag（失败/已存在冲突在前置校验拦截），随后交互询问
//         「推送到 origin / 仅保留本地」。
//
// 用法：
//   node scripts/release.mjs                    # 全交互（版本号留空时菜单选择）
//   node scripts/release.mjs v0.9.3             # 显式版本号（仍会询问是否推送）
//   node scripts/release.mjs --check            # 只跑门禁，不打 tag
//   node scripts/release.mjs v0.9.3 --dry-run   # 演练：门禁照跑，tag/推送只打印不执行
//   node scripts/release.mjs v0.9.3 --skip-test # 跳过 go test（gofmt/build/vet 仍执行）
//
// 口径说明：
//   - tag 为轻量 tag（git tag vX.Y.Z）；
//   - 同一 core 版本号（vX.Y.Z）只允许出现一次（正式版或某个预发布阶段）；
//   - 工作区有未提交改动时警告（不阻断）：tag 指向 HEAD，未提交改动不进入发版内容；
//   - 本地已存在同名 tag 且指向 HEAD → 视为上次未完成的发布，复用不重建；
//     指向其他提交 → 报错退出；
//   - 非交互终端（CI/管道）未显式给版本号时，列出候选后报错退出，不隐式取默认；
//     非交互终端跳过推送询问并打印手动推送命令。
//
// 依赖：Node ≥ 18、git、go（gofmt 随工具链自带）；不安装任何第三方包。

import readline from 'node:readline';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const ROOT = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
// 读取历史 tag 用：允许任意预发布后缀（候选计算时再区分）
const TAG_RE = /^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$/;
// 预发布阶段白名单：alpha / beta / rc，仅表示阶段，不带序号
const PRERELEASE_LABELS = ['alpha', 'beta', 'rc'];
// 用户输入的完整版本号（严格：拒绝前导零与 alpha.1 这类序号化预发布）
const STRICT_VERSION_RE = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-(alpha|beta|rc))?$/;
// 国内可达优先的 module proxy（可被环境变量 GOPROXY 覆盖）
const BUILD_GOPROXY = 'https://goproxy.cn,direct';
// 跨步骤共享的运行状态
const STATE = { dryRun: false, createdTag: '' };

// ---------- 基础工具 ----------

/**
 * 执行外部命令，输出直通终端，失败直接退出进程。
 *
 * @param {string} cmd 可执行文件名。
 * @param {string[]} args 参数列表。
 * @param {{cwd?: string, env?: Object, label?: string, allowFailure?: boolean}} [options]
 *   cwd=工作目录（默认 ROOT），env=追加环境变量，label=展示用命令行（默认自动拼接），
 *   allowFailure=true 时不退出仅返回 false。
 * @returns {boolean} 命令是否成功（allowFailure 时有意义）。
 * @throws {never} 命令不存在或失败（非 allowFailure）时打印原因并以退出码 1 结束。
 * @example
 * run('git', ['tag', 'v0.9.3', headCommit]);
 */
function run(cmd, args, options = {}) {
  const label = options.label ?? [cmd, ...args].join(' ');
  if (STATE.dryRun && options.dryRunSkip) {
    console.log(`  [dry-run] $ ${label}`);
    return true;
  }
  console.log(`  $ ${label}`);
  const result = spawnSync(cmd, args, {
    cwd: options.cwd ?? ROOT,
    env: { ...process.env, ...(options.env ?? {}) },
    stdio: ['ignore', 'inherit', 'inherit'],
  });
  if (result.error) {
    console.error(`错误: 无法执行 ${cmd}（请确认已安装且在 PATH 中）: ${result.error.message}`);
    process.exit(1);
  }
  if (result.status !== 0) {
    if (options.allowFailure) {
      console.error(`  ! 命令失败（退出码 ${result.status}）: ${label}`);
      return false;
    }
    console.error(`错误: 命令失败（退出码 ${result.status}）: ${label}`);
    process.exit(1);
  }
  return true;
}

/**
 * 执行外部命令并捕获标准输出（探测类命令，失败不退出）。
 *
 * @param {string} cmd 可执行文件名。
 * @param {string[]} args 参数列表。
 * @param {{cwd?: string, env?: Object}} [options] cwd=工作目录（默认 ROOT）。
 * @returns {{ok: boolean, stdout: string}} 成功标志与标准输出（已 trim）。
 * @example
 * probe('git', ['rev-parse', 'HEAD']).stdout; // 当前提交号
 */
function probe(cmd, args, options = {}) {
  const result = spawnSync(cmd, args, {
    cwd: options.cwd ?? ROOT,
    env: { ...process.env, ...(options.env ?? {}) },
    stdio: ['ignore', 'pipe', 'pipe'],
    encoding: 'utf8',
  });
  return {
    ok: !result.error && result.status === 0,
    stdout: (result.stdout ?? '').trim(),
  };
}

/**
 * 解析命令行参数。
 *
 * 支持 `--version vX.Y.Z` / `--version=vX.Y.Z` 与位置参数两种写法，
 * 以及 --check / --skip-test / --dry-run 开关。
 *
 * @param {string[]} argv process.argv.slice(2)。
 * @returns {{version: string, check: boolean, skipTest: boolean, dryRun: boolean}} 解析结果。
 * @throws {never} 未知参数时打印用法并以退出码 1 结束。
 * @example
 * parseArgs(['v0.9.3', '--dry-run']); // { version: 'v0.9.3', dryRun: true, ... }
 */
function parseArgs(argv) {
  const opts = { version: '', check: false, skipTest: false, dryRun: false };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === '--check') opts.check = true;
    else if (arg === '--skip-test') opts.skipTest = true;
    else if (arg === '--dry-run') opts.dryRun = true;
    else if (arg === '--version') opts.version = argv[++i] ?? '';
    else if (arg.startsWith('--version=')) opts.version = arg.slice('--version='.length);
    else if (arg === '--help' || arg === '-h') { printUsage(); process.exit(0); }
    else if (arg.startsWith('--')) { printUsage(); process.exit(1); }
    else opts.version = arg;
  }
  return opts;
}

/** 打印用法说明。 */
function printUsage() {
  console.log(`用法：node scripts/release.mjs [version] [选项]

  version             发布版本号（可省略 = 交互选择；形如 v0.9.3 或 v0.9.3-rc）
  --check             只跑门禁（gofmt/build/vet/test），不打 tag
  --skip-test         跳过 go test（gofmt/build/vet 仍执行）
  --dry-run           演练：门禁照跑，创建 tag 与推送只打印不执行`);
}

/**
 * 判断当前终端能否使用方向键菜单（raw mode）。
 *
 * CI、管道、部分受限终端返回 false，调用方应降级为编号输入或直接退出。
 *
 * @returns {boolean} 是否可用方向键菜单。
 * @example
 * if (!canUseRawMode()) { console.error('错误: 需要交互终端'); process.exit(1); }
 */
function canUseRawMode() {
  return Boolean(process.stdin.isTTY && process.stdout.isTTY
    && typeof process.stdin.setRawMode === 'function');
}

// 交互菜单临时切换 raw mode 时注册的恢复函数，避免非正常退出把终端留在 raw mode。
let rawModeRestore = null;

/**
 * 恢复终端状态（显示光标、退出 raw mode），可重复调用。
 *
 * @returns {void}
 * @example
 * restoreTerminal();
 */
function restoreTerminal() {
  if (!rawModeRestore) return;
  const restore = rawModeRestore;
  rawModeRestore = null;
  try {
    restore();
  } catch {
    // 终端已关闭时忽略
  }
}

/**
 * Ctrl+C / SIGINT 统一出口：恢复终端并退出（退出码 130）。
 *
 * 已创建但未推送的本地 tag 会保留，重跑同一版本号可复用。
 *
 * @param {string} [signal] 信号名，仅用于日志展示。
 * @returns {never} 直接退出进程。
 * @example
 * process.on('SIGINT', () => handleSignal('SIGINT'));
 */
function handleSignal(signal = 'SIGINT') {
  restoreTerminal();
  if (STATE.createdTag) {
    console.log(`\n已中断，本地 tag ${STATE.createdTag} 保留；重跑同一版本号可复用`);
  }
  process.exit(130);
}

/**
 * 方向键单选菜单（readline keypress + raw mode，无第三方依赖）。
 *
 * ↑↓ 移动、Enter 确认、Esc 或 q 取消；Ctrl+C 走 handleSignal 清理钩子。
 * 调用前应确认 canUseRawMode() 为 true。
 *
 * @param {string} title 菜单标题。
 * @param {string[]} items 选项文本（单行展示）。
 * @param {{initial?: number, hint?: string}} [options] 初始高亮项与底部提示。
 * @returns {Promise<number|null>} 选中项下标；Esc/q/取消返回 null。
 * @throws {never} Ctrl+C 时触发 handleSignal（退出码 130）。
 * @example
 * const idx = await selectMenu('选择版本方向', ['patch', 'minor']);
 */
function selectMenu(title, items, { initial = 0, hint = '↑/↓ 移动，Enter 确认，Esc/q 取消' } = {}) {
  return new Promise((resolve) => {
    const stdin = process.stdin;
    const out = process.stdout;
    const wasRaw = Boolean(stdin.isRaw);
    let index = Math.min(Math.max(initial, 0), items.length - 1);
    let drawn = 0;
    let settled = false;

    const lines = () => [
      title,
      ...items.map((text, i) => `  ${i === index ? '>' : ' '} ${text}`),
      `  ${hint}`,
    ];
    const draw = () => {
      if (drawn > 0) out.write(`\x1b[${drawn}A`);
      out.write('\x1b[J');
      out.write(`${lines().join('\n')}\n`);
      drawn = lines().length;
    };
    const cleanup = () => {
      stdin.removeListener('keypress', onKey);
      out.write('\x1b[?25h');
      try {
        stdin.setRawMode(wasRaw);
      } catch {
        // 终端不支持时忽略
      }
      stdin.pause();
      rawModeRestore = null;
    };
    const finish = (value) => {
      if (settled) return;
      settled = true;
      cleanup();
      resolve(value);
    };
    function onKey(_str, key) {
      if (!key) return;
      if (key.ctrl && key.name === 'c') {
        settled = true;
        cleanup();
        handleSignal('SIGINT');
        return;
      }
      if (key.name === 'escape' || key.name === 'q') {
        finish(null);
      } else if (key.name === 'up') {
        index = (index - 1 + items.length) % items.length;
        draw();
      } else if (key.name === 'down') {
        index = (index + 1) % items.length;
        draw();
      } else if (key.name === 'return' || key.name === 'enter') {
        finish(index);
      }
    }

    rawModeRestore = () => {
      out.write('\x1b[?25h');
      stdin.setRawMode(wasRaw);
    };
    readline.emitKeypressEvents(stdin);
    stdin.setRawMode(true);
    stdin.resume();
    out.write('\x1b[?25l');
    draw();
    stdin.on('keypress', onKey);
  });
}

/**
 * 单次逐行提问（临时创建 readline；用于菜单中「手动输入完整版本号」）。
 *
 * @param {string} question 提示语。
 * @returns {Promise<string|undefined>} 用户输入；stdin 关闭时返回 undefined。
 * @example
 * const raw = await askLineOnce('  请输入完整版本号: ');
 */
function askLineOnce(question) {
  return new Promise((resolve) => {
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
    rl.once('close', () => resolve(undefined));
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

// ---------- 版本解析与候选 ----------

/**
 * 版本号写法规范化：允许省略 v 前缀，统一补齐。
 *
 * @param {string} raw 用户原始输入。
 * @returns {string} 以 v 开头的版本号。
 * @example
 * normalizeVersion('0.9.3'); // 'v0.9.3'
 */
function normalizeVersion(raw) {
  const text = String(raw ?? '').trim();
  return text && !text.startsWith('v') ? `v${text}` : text;
}

/**
 * 解析版本 tag 为结构化对象（历史 tag 解析不受 STRICT_VERSION_RE 限制）。
 *
 * @param {string} tag 形如 v0.9.3 / v0.9.3-rc / v0.9.3-alpha.1 的 tag。
 * @returns {Object|null} { tag, core, major, minor, patch, stage, num }；不合法返回 null。
 * @example
 * parseVersionTag('v0.9.3-rc.1').stage; // 'rc'
 */
function parseVersionTag(tag) {
  const m = /^v(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$/.exec(tag);
  if (!m) return null;
  const suffix = m[4] ?? '';
  const legacy = /^([A-Za-z]+)\.(\d+)$/.exec(suffix);
  return {
    tag,
    core: `${Number(m[1])}.${Number(m[2])}.${Number(m[3])}`,
    major: Number(m[1]),
    minor: Number(m[2]),
    patch: Number(m[3]),
    stage: suffix === '' ? '' : legacy ? legacy[1] : suffix,
    num: legacy ? Number(legacy[2]) : 0,
  };
}

/**
 * 预发布阶段权重：alpha < beta < rc < 正式版；未知后缀按最低权重处理。
 *
 * @param {string} stage 阶段标识（空串表示正式版）。
 * @returns {number} 权重值（越大越新）。
 * @example
 * versionRank('rc'); // 3
 */
function versionRank(stage) {
  if (!stage) return PRERELEASE_LABELS.length + 1;
  const index = PRERELEASE_LABELS.indexOf(stage);
  return index >= 0 ? index + 1 : 0;
}

/**
 * 纯 JS 版本比较（替代 git version sort，跨平台/跨机器结果一致）。
 *
 * 先比较 major/minor/patch，再比较阶段权重，最后比较旧格式的预发布序号。
 *
 * @param {Object} a parseVersionTag 的结果。
 * @param {Object} b parseVersionTag 的结果。
 * @returns {number} 负数 = a 更旧，正数 = a 更新，0 = 相同。
 * @example
 * compareVersion(parseVersionTag('v0.9.2'), parseVersionTag('v0.9.3')); // 负数
 */
function compareVersion(a, b) {
  const core = (a.major - b.major) || (a.minor - b.minor) || (a.patch - b.patch);
  if (core !== 0) return core;
  const rank = versionRank(a.stage) - versionRank(b.stage);
  if (rank !== 0) return rank;
  return a.num - b.num;
}

/**
 * 取全部 tag 中语义最高的版本。
 *
 * @param {string[]} allTags 全部合法 v* tag。
 * @returns {Object|null} parseVersionTag 结果；无合法 tag 返回 null。
 * @example
 * latestParsedVersion(['v0.9.1', 'v0.9.2']).tag; // 'v0.9.2'
 */
function latestParsedVersion(allTags) {
  return allTags
    .map((tag) => parseVersionTag(tag))
    .filter(Boolean)
    .reduce((max, cur) => (max && compareVersion(cur, max) < 0 ? max : cur), null);
}

/**
 * 读取本地全部合法 v* 发布 tag。
 *
 * @returns {string[]} tag 列表（按字母序，未排序展示用）。
 * @throws {never} git 执行失败时以退出码 1 结束。
 * @example
 * listLocalTags(); // ['v0.8.0', 'v0.9.2', ...]
 */
function listLocalTags() {
  const res = probe('git', ['tag', '--list', 'v*']);
  if (!res.ok && res.stdout === '' && probe('git', ['rev-parse', '--git-dir']).ok === false) {
    console.error('错误: 当前目录不是 git 仓库（脚本需在 jcbaseGo 仓库内运行）');
    process.exit(1);
  }
  return res.stdout.split('\n').map((t) => t.trim()).filter((t) => TAG_RE.test(t));
}

/**
 * 按最新版本生成三个方向的正式版候选。
 *
 * 涨哪一位由变更内容决定：修 bug → patch，新增功能 → minor，不兼容/大改 → major。
 *
 * @param {string[]} allTags 全部合法 v* tag。
 * @returns {Array<{version: string, kind: string, label: string}>} 三个方向的候选。
 * @example
 * buildVersionCandidates(['v0.9.2'])[0].version; // 'v0.9.3'
 */
function buildVersionCandidates(allTags) {
  const latest = latestParsedVersion(allTags);
  if (!latest) {
    return [
      { version: 'v0.1.0', kind: 'patch', label: '首个版本（修 bug 级别）' },
      { version: 'v0.2.0', kind: 'minor', label: '首个版本（minor 起点）' },
      { version: 'v1.0.0', kind: 'major', label: '首个版本（major 起点）' },
    ];
  }
  const base = latest.tag;
  return [
    { version: `v${latest.major}.${latest.minor}.${latest.patch + 1}`, kind: 'patch', label: `修 bug（基于 ${base}，patch +1）` },
    { version: `v${latest.major}.${latest.minor + 1}.0`, kind: 'minor', label: `新增功能（基于 ${base}，minor +1）` },
    { version: `v${latest.major + 1}.0.0`, kind: 'major', label: `不兼容/大改（基于 ${base}，major +1）` },
  ];
}

/**
 * 找出与目标版本共用同一 core（vX.Y.Z）的其它 tag。
 *
 * 规则：同一版本号只会出现一次（正式版或某个预发布阶段），
 * 因此 core 相同、tag 名不同的记录都算冲突。
 *
 * @param {string[]} allTags 全部合法 v* tag。
 * @param {string} version 目标版本号。
 * @returns {string[]} 同 core 的其它 tag。
 * @example
 * sameCoreTags(['v0.9.3-rc'], 'v0.9.3'); // ['v0.9.3-rc']
 */
function sameCoreTags(allTags, version) {
  const target = parseVersionTag(version);
  if (!target) return [];
  return allTags
    .map((tag) => parseVersionTag(tag))
    .filter((p) => p && p.core === target.core && p.tag !== version)
    .map((p) => p.tag);
}

/**
 * 打印完整候选清单（三个方向 × 四个阶段 = 12 项）。
 *
 * @param {string[]} allTags 全部合法 v* tag（用于展示当前最高版本）。
 * @returns {void}
 * @example
 * printCandidateCatalog(listLocalTags());
 */
function printCandidateCatalog(allTags) {
  const latest = latestParsedVersion(allTags);
  console.log(`  当前最高版本：${latest ? latest.tag : '无（仓库暂无 v* tag）'}`);
  console.log('  版本候选（3 个方向 × 4 个阶段 = 12 项；预发布不带序号，只在版本号后加 -alpha/-beta/-rc）：');
  buildVersionCandidates(allTags).forEach((direction, i) => {
    const mark = allTags.includes(direction.version) ? '  [已存在]' : '';
    const recommend = i === 0 ? '  ← 默认推荐' : '';
    console.log(`    ${i + 1}) ${direction.version}  正式版  ${direction.label}${mark}${recommend}`);
    PRERELEASE_LABELS.forEach((stage) => {
      const version = `${direction.version}-${stage}`;
      const preMark = allTags.includes(version) ? '  [已存在]' : '';
      console.log(`       ${version}  ${stage.padEnd(5)} 预发布${preMark}`);
    });
  });
  console.log(`  选择写法：1-3 取正式版；1-alpha、2-beta、3-rc 取预发布；也可直接输入完整版本号（如 ${buildVersionCandidates(allTags)[0].version}-rc）`);
}

/**
 * 严格校验用户输入的版本号（历史 tag 解析不受此限制）。
 *
 * 只接受 vX.Y.Z 或 vX.Y.Z-alpha/-beta/-rc；拒绝前导零、未知后缀、
 * 以及旧版序号化预发布（alpha.1）。
 *
 * @param {string} version 完整版本号（应已补 v 前缀）。
 * @returns {string} 错误原因；合法时返回空串。
 * @example
 * validateVersion('v0.9.3-rc.1'); // '预发布已取消序号...'
 */
function validateVersion(version) {
  if (/-[A-Za-z]+\.\d+$/.test(version)) {
    return '预发布已取消序号：请写 v0.9.3-alpha / -beta / -rc（不要再带 .1）';
  }
  if (!STRICT_VERSION_RE.test(version)) {
    return `版本号格式不合法 "${version}"：需形如 v0.9.3 或 v0.9.3-alpha/-beta/-rc（不接受前导零或未知后缀）`;
  }
  return '';
}

// ---------- 流程步骤 ----------

/**
 * 步骤 1：解析发布版本号（显式传入校验，否则交互两步菜单选择）。
 *
 * @param {Object} opts parseArgs 的结果。
 * @param {string[]} allTags 全部合法 v* tag。
 * @returns {Promise<string>} 最终版本号（v 开头）。
 * @throws {never} 显式版本非法、冲突，或非交互终端未给版本时以退出码 1 结束。
 * @example
 * const version = await resolveVersion(opts, listLocalTags());
 */
async function resolveVersion(opts, allTags) {
  if (opts.version) {
    const version = normalizeVersion(opts.version);
    const error = validateVersion(version);
    if (error) {
      console.error(`错误: ${error}`);
      process.exit(1);
    }
    return version;
  }
  if (!canUseRawMode()) {
    console.error('错误: 未指定发布版本（CI / 非 TTY 下不会自动取默认）');
    printCandidateCatalog(allTags);
    console.error('  可直接复制：node scripts/release.mjs --version <上面任一候选>');
    process.exit(1);
  }
  return await promptVersion(allTags);
}

/**
 * 交互两步选择版本：第一步选变更方向（patch/minor/major/手动输入），
 * 第二步选阶段（正式/alpha/beta/rc）。
 *
 * @param {string[]} allTags 全部合法 v* tag。
 * @returns {Promise<string>} 最终版本号（v 开头）；取消选择时以退出码 0 结束。
 * @throws {never} 手动输入的版本非法时打印原因并回到输入（Ctrl+C 退出 130）。
 * @example
 * const version = await promptVersion(listLocalTags());
 */
async function promptVersion(allTags) {
  printCandidateCatalog(allTags);
  const directions = buildVersionCandidates(allTags);
  const items = directions.map((d) => `${d.version}  ${d.label}`);
  items.push('手动输入完整版本号');

  // eslint-disable-next-line no-constant-condition
  while (true) {
    const idx = await selectMenu('第一步：选择版本方向', items, { initial: 0 });
    if (idx === null) {
      console.log('已取消，未做任何变更');
      process.exit(0);
    }
    if (idx === directions.length) {
      const raw = await askLineOnce('  请输入完整版本号（如 v0.9.3-rc）: ');
      const version = normalizeVersion(raw);
      const error = validateVersion(version);
      if (error) {
        console.error(`  ${error}`);
        continue;
      }
      return version;
    }
    const stageIdx = await selectMenu('第二步：选择阶段', ['正式版', '-alpha', '-beta', '-rc'], { initial: 0 });
    if (stageIdx === null) continue;
    const stage = ['', 'alpha', 'beta', 'rc'][stageIdx];
    return stage ? `${directions[idx].version}-${stage}` : directions[idx].version;
  }
}

/**
 * 步骤 2：轻量门禁（gofmt -l . + go build/vet/test ./...，CI 同口径）。
 *
 * --skip-test 只跳过 go test（gofmt/build/vet 仍执行，避免「跳过测试连
 * 编译检查也一起跳过」的副作用）；任一步失败即中止，后续不打 tag。
 *
 * @param {Object} opts parseArgs 的结果（skipTest）。
 * @returns {void}
 * @throws {never} 任一步失败时打印原因并以退出码 1 结束。
 * @example
 * runGate({ skipTest: false });
 */
function runGate(opts) {
  console.log('==> [2/3] 轻量门禁（gofmt + go build/vet/test）');
  const goEnv = { GOPROXY: process.env.GOPROXY || BUILD_GOPROXY };

  const fmt = probe('gofmt', ['-l', '.']);
  if (!fmt.ok) {
    console.error('错误: gofmt 执行失败（请确认 go 工具链在 PATH 中）');
    process.exit(1);
  }
  if (fmt.stdout !== '') {
    console.error('错误: 以下文件未通过 gofmt 格式检查：');
    console.error(fmt.stdout);
    process.exit(1);
  }
  console.log('  gofmt 通过');

  run('go', ['build', './...'], { env: goEnv });
  run('go', ['vet', './...'], { env: goEnv });
  if (opts.skipTest) {
    console.warn('警告: --skip-test 跳过 go test（gofmt/build/vet 已执行；库项目不建议跳过测试）');
    return;
  }
  run('go', ['test', './...'], { env: goEnv });
}

/**
 * 步骤 3 前置校验：版本冲突、同名 tag 复用判定与脏工作区警告。
 *
 * 校验顺序：同 core 其它 tag（冲突）→ 本地同名 tag（指向 HEAD 可复用，
 * 指向其他提交报错）→ 未提交改动警告（不阻断）。返回是否复用已有 tag。
 *
 * @param {string} version 最终版本号。
 * @param {string[]} allTags 全部合法 v* tag。
 * @returns {boolean} true = 同名 tag 已存在且指向 HEAD，跳过创建。
 * @throws {never} 同 core 冲突或同名 tag 指向其他提交时以退出码 1 结束。
 * @example
 * const reuse = preflightTag('v0.9.3', listLocalTags());
 */
function preflightTag(version, allTags) {
  const conflicts = sameCoreTags(allTags, version);
  if (conflicts.length > 0) {
    console.error(`错误: 版本 ${version} 与已有 tag 冲突（同一版本号只会出现一次）：${conflicts.join('、')}`);
    console.error('  请换一个版本号（正式版必须使用下一个版本号，不存在「把 -rc 提成同号正式版」的路径）');
    process.exit(1);
  }

  const headCommit = probe('git', ['rev-parse', 'HEAD']).stdout;
  const tagCommit = probe('git', ['rev-parse', '--verify', `refs/tags/${version}^{commit}`]).stdout;
  const reuse = Boolean(tagCommit);
  if (reuse && tagCommit !== headCommit) {
    console.error(`错误: 本地 tag ${version} 已存在且指向 ${tagCommit.slice(0, 12)}（HEAD 为 ${headCommit.slice(0, 12)}）`);
    console.error('  不允许复用指向其他提交的同名 tag；如需重建请先 git tag -d ' + version);
    process.exit(1);
  }

  const dirty = probe('git', ['status', '--porcelain']).stdout;
  if (dirty) {
    console.warn('警告: 工作区有未提交改动，tag 指向 HEAD，以下改动不进入本次发版内容：');
    console.warn(dirty.split('\n').map((line) => `    ${line}`).join('\n'));
  }
  return reuse;
}

/**
 * 步骤 3：创建轻量 tag（git tag vX.Y.Z，指向 HEAD）。
 *
 * 已存在且指向 HEAD 的同名 tag（上次未完成的发布）跳过创建直接复用；
 * dry-run 只打印命令。创建结果记入 STATE，中断时保留现场提示续跑。
 *
 * @param {string} version 最终版本号。
 * @param {boolean} reuse 前置校验返回的复用标志。
 * @returns {void}
 * @throws {never} tag 创建失败时打印原因并以退出码 1 结束。
 * @example
 * createTag('v0.9.3', false);
 */
function createTag(version, reuse) {
  const headCommit = probe('git', ['rev-parse', 'HEAD']).stdout;
  if (reuse) {
    console.log(`==> [3/3] 本地 tag ${version} 已存在且指向 HEAD ${headCommit.slice(0, 12)}，复用（上次未完成的发布）`);
    STATE.createdTag = version;
    return;
  }
  console.log('==> [3/3] 创建轻量 tag');
  if (STATE.dryRun) {
    console.log(`  [dry-run] $ git tag ${version} ${headCommit.slice(0, 12)}（实际执行时创建轻量 tag）`);
    return;
  }
  const ok = run('git', ['tag', version, headCommit], { allowFailure: true });
  if (!ok) {
    console.error(`错误: 创建轻量 tag ${version} 失败（可能已存在同名 tag 或无权限）`);
    process.exit(1);
  }
  STATE.createdTag = version;
  console.log(`  已创建轻量 tag: ${version} → ${headCommit.slice(0, 12)}（中断时保留，重跑同一版本号可复用）`);
}

/**
 * 步骤 3 收尾：交互询问是否推送 tag 到 origin。
 *
 * 推送只推本次 tag（git push origin <tag>），不推分支；dry-run 只打印；
 * 非交互终端跳过询问并打印手动推送命令；取消推送时本地 tag 保留。
 *
 * @param {string} version 最终版本号。
 * @returns {Promise<void>}
 * @throws {never} 推送失败时打印可重试命令并以退出码 1 结束。
 * @example
 * await maybePush('v0.9.3');
 */
async function maybePush(version) {
  if (STATE.dryRun) {
    console.log(`  [dry-run] 此处将交互询问是否执行：git push origin ${version}`);
    return;
  }
  if (!canUseRawMode()) {
    console.log(`非交互终端，跳过推送询问；需要时手动执行：git push origin ${version}`);
    return;
  }
  const idx = await selectMenu(`是否推送 ${version} 到 origin？`, [
    `推送到 origin（git push origin ${version}）`,
    '仅保留本地（稍后可手动 git push origin ' + version + '）',
  ], { initial: 1, hint: '↑/↓ 移动，Enter 确认，Esc/q 取消（取消 = 仅保留本地）' });
  if (idx !== 0) {
    console.log(`已保留本地 tag ${version}，未推送`);
    return;
  }
  const ok = run('git', ['push', 'origin', version], { allowFailure: true });
  if (!ok) {
    console.error(`错误: 推送失败，本地 tag ${version} 已保留；可稍后重试：git push origin ${version}`);
    process.exit(1);
  }
  console.log(`已推送 ${version} 到 origin（Go module proxy 拉取新版本可能需要几分钟同步）`);
}

/**
 * 脚本主入口：版本选择 → 前置校验 → 门禁 → 打 tag → 询问推送。
 *
 * @returns {Promise<void>}
 * @example
 * // 终端执行：node scripts/release.mjs
 */
async function main() {
  const opts = parseArgs(process.argv.slice(2));
  STATE.dryRun = opts.dryRun;
  if (opts.dryRun) console.log('==> dry-run 演练模式：tag 与推送只打印不执行\n');

  if (opts.check) {
    console.log('==> 门禁检查模式（--check 只读，不创建 tag）');
    runGate(opts);
    console.log('==> 门禁通过');
    return;
  }

  const allTags = listLocalTags();
  console.log('==> [1/3] 选择发布版本');
  const version = await resolveVersion(opts, allTags);
  console.log(`  发布版本：${version}`);
  const reuse = preflightTag(version, allTags);
  runGate(opts);
  createTag(version, reuse);
  await maybePush(version);
  console.log('==> 发版完成');
}

process.on('SIGINT', () => handleSignal('SIGINT'));
process.on('SIGTERM', () => handleSignal('SIGTERM'));
main();
