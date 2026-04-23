const fs = require("fs");
const path = require("path");

const PLATFORM_MAP = {
  "linux-x64": "@physics91org/linux-x64",
  "linux-arm64": "@physics91org/linux-arm64",
  "darwin-x64": "@physics91org/darwin-x64",
  "darwin-arm64": "@physics91org/darwin-arm64",
  "win32-x64": "@physics91org/win32-x64",
};

function getBinaryExtension(platform = process.platform) {
  return platform === "win32" ? ".exe" : "";
}

function getPlatformPackage(platform = process.platform, arch = process.arch) {
  return PLATFORM_MAP[`${platform}-${arch}`] || null;
}

function getSupportedPlatforms() {
  return Object.keys(PLATFORM_MAP);
}

function getSidecarBinaryPath(baseDir, platform = process.platform) {
  return path.join(baseDir, `naverworks-bin${getBinaryExtension(platform)}`);
}

function resolvePlatformBinaryPath(baseDir, platform = process.platform, arch = process.arch) {
  const packageName = getPlatformPackage(platform, arch);
  if (!packageName) {
    return { packageName: null, binaryPath: null };
  }

  try {
    const packageJsonPath = require.resolve(`${packageName}/package.json`, {
      paths: [baseDir],
    });
    const packageDir = path.dirname(packageJsonPath);
    const binaryPath = path.join(packageDir, `naverworks${getBinaryExtension(platform)}`);

    if (fs.existsSync(binaryPath)) {
      return { packageName, binaryPath };
    }
  } catch {}

  return { packageName, binaryPath: null };
}

module.exports = {
  getPlatformPackage,
  getSupportedPlatforms,
  getSidecarBinaryPath,
  resolvePlatformBinaryPath,
};
