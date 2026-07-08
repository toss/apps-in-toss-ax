#!/usr/bin/env node

import { execSync } from "node:child_process";
import fs from "node:fs";

const SEMVER_RE =
  /^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z0-9.-]+))?(?:\+([0-9A-Za-z0-9.-]+))?$/;

function error(message) {
  console.error(`::error::${message}`);
  process.exit(1);
}

function writeOutput(key, value) {
  const file = process.env.GITHUB_OUTPUT;
  if (file) {
    fs.appendFileSync(file, `${key}=${value}\n`);
  }
}

function readPackageVersion() {
  const pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));
  return pkg.version;
}

function parseSemver(version) {
  const match = version.match(SEMVER_RE);
  if (!match) {
    return null;
  }

  return {
    major: Number(match[1]),
    minor: Number(match[2]),
    patch: Number(match[3]),
    prerelease: match[4] ?? null,
  };
}

function compareSemver(a, b) {
  const left = parseSemver(a);
  const right = parseSemver(b);
  if (!left || !right) {
    throw new Error(`Invalid semver comparison: ${a} vs ${b}`);
  }

  for (const key of ["major", "minor", "patch"]) {
    if (left[key] !== right[key]) {
      return left[key] - right[key];
    }
  }

  if (left.prerelease === null && right.prerelease === null) {
    return 0;
  }
  if (left.prerelease === null) {
    return 1;
  }
  if (right.prerelease === null) {
    return -1;
  }
  return left.prerelease.localeCompare(right.prerelease);
}

function maxSemver(...versions) {
  const valid = versions.filter((version) => version && version !== "null" && parseSemver(version));
  if (valid.length === 0) {
    return "";
  }

  return valid.sort(compareSemver).at(-1);
}

function validateVersion(version, ...minimumVersions) {
  if (!version || version === "null") {
    error("version is missing");
  }

  if (version.startsWith("v")) {
    error(`Version must not start with 'v' (use 1.0.0, not v1.0.0): ${version}`);
  }

  if (!parseSemver(version)) {
    error(`Invalid semver format: ${version}`);
  }

  const baseline = maxSemver(...minimumVersions);
  if (baseline && compareSemver(version, baseline) <= 0) {
    error(`Version ${version} must be greater than ${baseline}`);
  }

  try {
    execSync(`git rev-parse v${version}`, { stdio: "ignore" });
    error(`Git tag v${version} already exists`);
  } catch {
    // tag does not exist
  }
}

function readPreviousPackageVersion(before) {
  try {
    execSync(`git cat-file -e ${before}:package.json`, { stdio: "ignore" });
  } catch {
    return null;
  }

  const contents = execSync(`git show ${before}:package.json`, {
    encoding: "utf8",
  });
  return JSON.parse(contents).version;
}

function latestTagVersion() {
  const output = execSync("git tag -l 'v*'", { encoding: "utf8" }).trim();
  if (!output) {
    return "";
  }

  return output
    .split("\n")
    .map((tag) => tag.replace(/^v/, ""))
    .sort(compareSemver)
    .at(-1);
}

function resolveManualRelease() {
  const inputVersion = process.env.INPUT_VERSION;
  if (!inputVersion) {
    error("INPUT_VERSION is required for workflow_dispatch");
  }

  const pkgVersion = readPackageVersion();
  if (inputVersion !== pkgVersion) {
    error(
      `Input version (${inputVersion}) does not match package.json (${pkgVersion})`
    );
  }

  const latestTag = latestTagVersion();
  validateVersion(pkgVersion, latestTag);
  writeOutput("version", pkgVersion);
  writeOutput("should_release", "true");
  console.info(`Manual release: v${pkgVersion}`);
}

function resolvePushRelease() {
  const before = process.env.EVENT_BEFORE;
  if (!before) {
    error("EVENT_BEFORE is required for push");
  }

  const current = readPackageVersion();
  writeOutput("version", current);

  if (before === "0000000000000000000000000000000000000000") {
    writeOutput("should_release", "false");
    console.info("Skipping: initial push to branch");
    return;
  }

  const latestTag = latestTagVersion();
  const previous = readPreviousPackageVersion(before);
  if (previous === null) {
    validateVersion(current, latestTag);
    writeOutput("should_release", "true");
    console.info("No previous package.json found; treating as version change");
    return;
  }

  if (current !== previous) {
    validateVersion(current, previous, latestTag);
    writeOutput("should_release", "true");
    console.info(`Version changed: ${previous} -> ${current}`);
    return;
  }

  writeOutput("should_release", "false");
  console.info(`package.json changed but version unchanged: ${current}`);
}

function main() {
  switch (process.env.EVENT_NAME ?? "push") {
    case "workflow_dispatch":
      resolveManualRelease();
      break;
    case "push":
      resolvePushRelease();
      break;
    default:
      error(`Unsupported EVENT_NAME: ${process.env.EVENT_NAME}`);
  }
}

main();
