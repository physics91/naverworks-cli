const fs = require("fs");
const path = require("path");

const PLATFORM_MAP = {
  "linux-x64": "@nw-cli/linux-x64",
  "linux-arm64": "@nw-cli/linux-arm64",
  "darwin-x64": "@nw-cli/darwin-x64",
  "darwin-arm64": "@nw-cli/darwin-arm64",
  "win32-x64": "@nw-cli/win32-x64",
};

function getPlatformPackage() {
  const key = `${process.platform}-${process.arch}`;
  const pkg = PLATFORM_MAP[key];
  if (!pkg) {
    console.error(
      `지원하지 않는 플랫폼입니다: ${process.platform}-${process.arch}`
    );
    console.error(`지원 플랫폼: ${Object.keys(PLATFORM_MAP).join(", ")}`);
    process.exit(1);
  }
  return pkg;
}

function getBinaryPath(pkg) {
  const ext = process.platform === "win32" ? ".exe" : "";
  try {
    const pkgDir = path.dirname(require.resolve(`${pkg}/package.json`));
    return path.join(pkgDir, `nw-cli${ext}`);
  } catch {
    return null;
  }
}

function install() {
  const pkg = getPlatformPackage();
  const binaryPath = getBinaryPath(pkg);

  if (!binaryPath || !fs.existsSync(binaryPath)) {
    console.error(`바이너리를 찾을 수 없습니다: ${pkg}`);
    console.error("npm install을 다시 실행해주세요.");
    process.exit(1);
  }

  const binDir = path.join(__dirname, "bin");
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  const ext = process.platform === "win32" ? ".exe" : "";
  const dest = path.join(binDir, `nw-cli${ext}`);

  fs.copyFileSync(binaryPath, dest);
  fs.chmodSync(dest, 0o755);
}

install();
