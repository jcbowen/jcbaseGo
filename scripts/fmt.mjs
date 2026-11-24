#!/usr/bin/env node
// jcbaseGo Go 代码格式化（跨平台唯一入口，对齐 market/scripts/fmt.mjs）。
//
// 依赖：Node ≥18 + go 工具链自带的 gofmt（不安装任何第三方包）。
// 默认范围与根目录 go.mod 的模块代码目录一致：component errcode example middleware
// （scripts 只放 Node 脚本，不在格式化范围）。
//
// 用法：
//   node scripts/fmt.mjs                    # 自动修复（gofmt -w，幂等）
//   node scripts/fmt.mjs --check            # 只检查：列出未格式化文件与 diff，退出码 1
//   node scripts/fmt.mjs --check component  # 只检查指定目录/文件（可传多个）
//   node scripts/fmt.mjs --help             # 查看帮助

import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const DEFAULT_TARGETS = ['component', 'errcode', 'example', 'middleware'];

/**
 * 打印脚本用法。
 *
 * @returns {void}
 * @example
 * usage();
 */
function usage() {
  console.log(`用法：node scripts/fmt.mjs [--check] [目标...]

  --check, -c   只检查不写文件；未通过时列出文件与 diff 并退出 1
  --fix         自动修复（默认行为）
  -h, --help    显示本帮助

目标缺省为 component errcode example middleware（根目录下全部 Go 代码目录，相对路径）。`);
}

/**
 * 执行 gofmt 子命令并返回标准输出。
 *
 * @param {string[]} args gofmt 参数列表。
 * @returns {string} gofmt 的标准输出（保留换行，由调用方 trim）。
 * @throws {never} 执行失败时直接打印原因并以退出码 1 结束进程。
 * @example
 * runGofmt(['-l', 'component', 'middleware']);
 */
function runGofmt(args) {
  try {
    return execFileSync('gofmt', args, {
      cwd: ROOT,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
  } catch (err) {
    const stderr = err?.stderr ? String(err.stderr).trim() : '';
    console.error(`错误: gofmt ${args.join(' ')} 执行失败（请确认 go 在 PATH 中）: ${err.message}`);
    if (stderr) console.error(stderr);
    process.exit(1);
  }
}

/**
 * 格式检查：列出不符合 gofmt 的文件与 diff。
 *
 * @param {string[]} targets 目标目录或文件列表。
 * @returns {void} 检查通过返回 0，未通过退出码 1。
 * @example
 * check(['component']);
 */
function check(targets) {
  const unformatted = runGofmt(['-l', ...targets]).trim();
  if (!unformatted) {
    console.log(`gofmt 检查通过（${targets.join(' ')}）`);
    return;
  }

  console.error(`以下文件未通过 gofmt（运行 node scripts/fmt.mjs 自动修复）：`);
  for (const file of unformatted.split('\n').filter(Boolean)) console.error(`  ${file}`);
  console.error('--- gofmt diff ---');
  console.error(runGofmt(['-d', ...targets]).trimEnd());
  process.exit(1);
}

/**
 * 自动修复：对目标执行 gofmt -w。
 *
 * @param {string[]} targets 目标目录或文件列表。
 * @returns {void}
 * @example
 * fix(['component', 'errcode', 'example', 'middleware']);
 */
function fix(targets) {
  const changed = runGofmt(['-l', '-w', ...targets]).trim();
  if (!changed) {
    console.log(`gofmt 无需修复（${targets.join(' ')}）`);
    return;
  }

  console.log('已格式化以下文件：');
  for (const file of changed.split('\n').filter(Boolean)) console.log(`  ${file}`);
}

/**
 * 解析命令行参数。
 *
 * @param {string[]} argv 原始参数（process.argv.slice(2)）。
 * @returns {{mode: 'check'|'fix', targets: string[]}} 解析结果。
 * @throws {never} 参数错误时打印帮助并以退出码 1 结束进程。
 * @example
 * parseArgs(['--check', 'component']);
 */
function parseArgs(argv) {
  let mode = 'fix';
  const targets = [];

  for (const arg of argv) {
    if (arg === '--check' || arg === '-c') {
      mode = 'check';
    } else if (arg === '--fix') {
      mode = 'fix';
    } else if (arg === '-h' || arg === '--help') {
      usage();
      process.exit(0);
    } else if (arg.startsWith('-')) {
      console.error(`未知参数: ${arg}`);
      usage();
      process.exit(1);
    } else {
      targets.push(arg);
    }
  }

  return { mode, targets: targets.length > 0 ? targets : DEFAULT_TARGETS };
}

const { mode, targets } = parseArgs(process.argv.slice(2));
if (mode === 'check') {
  check(targets);
} else {
  fix(targets);
}
