const fs = require("fs");
const path = require("path");
const {
  getPlatformPackage,
  getSupportedPlatforms,
  getSidecarBinaryPath,
  resolvePlatformBinaryPath,
} = require("./platform");

function install() {
  const pkg = getPlatformPackage();
  if (!pkg) {
    console.error(
      `지원하지 않는 플랫폼입니다: ${process.platform}-${process.arch}`
    );
    console.error(`지원 플랫폼: ${getSupportedPlatforms().join(", ")}`);
    process.exit(1);
  }

  const { binaryPath } = resolvePlatformBinaryPath(__dirname);

  if (!binaryPath || !fs.existsSync(binaryPath)) {
    console.error(`바이너리를 찾을 수 없습니다: ${pkg}`);
    console.error("npm 또는 bun으로 패키지를 다시 설치해주세요.");
    process.exit(1);
  }

  const binDir = path.join(__dirname, "bin");
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  const dest = getSidecarBinaryPath(binDir);

  fs.copyFileSync(binaryPath, dest);
  fs.chmodSync(dest, 0o755);
}

install();
