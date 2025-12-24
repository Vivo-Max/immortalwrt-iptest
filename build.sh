#!/usr/bin/env bash
set -e
set -o pipefail
# =========================================================
# 1. 基础配置
# =========================================================
# 源码在仓库中的相对路径
CUSTOM_PACKAGE_DIR="src/package/custom"
# 平台定义 (索引 1 是 RK3588)
PLATFORMS=(
  "mediatek/filogic/mediatek_filogic_DEVICE_cmcc_rax3000m"
  "rockchip/rk3588/orangepi5plus"
  "x86/64/generic"
)
IMMORTALWRT_REPO="https://github.com/immortalwrt/immortalwrt.git"
# =========================================================
# 2. 构建准备函数
# =========================================================
prepare_build_dir() {
    local BUILD_DIR="$1"
    local PLATFORM_CFG="$2"
    local FMT="$3"
    local ROOT_DIR=$(cd "$(dirname "$0")"; pwd)
    local TARGET=$(echo "$PLATFORM_CFG" | cut -d'/' -f1)
    local SUBTARGET=$(echo "$PLATFORM_CFG" | cut -d'/' -f2)
    local DEVICE=$(echo "$PLATFORM_CFG" | cut -d'/' -f3)
    echo "==> 正在准备环境: $BUILD_DIR"
    # 拉取源码 (GitHub 环境下直接克隆极快)
    if [ -d "$BUILD_DIR" ]; then
        pushd "$BUILD_DIR" >/dev/null
        git pull origin master || true
        popd >/dev/null
    else
        git clone --depth 1 "$IMMORTALWRT_REPO" "$BUILD_DIR"
    fi
    pushd "$BUILD_DIR" >/dev/null
   
    # 更新 Feeds
    ./scripts/feeds update -a
    ./scripts/feeds install -a
    # --- 核心修复：物理清理冲突包 (解除递归依赖) ---
    echo "--> 移除冲突包索引..."
    rm -rf "package/feeds/packages/backuppc"
    rm -rf "package/feeds/packages/lxc"
    rm -rf "package/feeds/luci/luci-app-lxc"
    rm -rf "package/feeds/packages/oci-runtime-tests"
    rm -rf "tmp"
    # --- 同步自定义源码 ---
    echo "--> 同步本地源码至构建目录..."
    mkdir -p package/custom
    if [ -d "$ROOT_DIR/$CUSTOM_PACKAGE_DIR" ]; then
        rsync -a --exclude ".git" "$ROOT_DIR/$CUSTOM_PACKAGE_DIR/" "package/custom/"
    else
        echo "错误: 找不到目录 $CUSTOM_PACKAGE_DIR"
        exit 1
    fi
    # 写入 .config 配置文件
    cat > .config << EOF
CONFIG_TARGET_${TARGET}=y
CONFIG_TARGET_${TARGET}_${SUBTARGET}=y
CONFIG_TARGET_DEVICE_${DEVICE}=y
CONFIG_PACKAGE_backuppc=n
CONFIG_PACKAGE_lxc=n
CONFIG_PACKAGE_oci-runtime-tests=n
CONFIG_PACKAGE_tar=y
CONFIG_PACKAGE_TAR_XZ=y
CONFIG_ALL_KMODS=y
CONFIG_PACKAGE_luci=y
CONFIG_PACKAGE_iptest=m
CONFIG_PACKAGE_luci-app-iptest=m
EOF
   
    [ "$FMT" = "apk" ] && echo "CONFIG_PACKAGE_APK=y" >> .config || echo "CONFIG_PACKAGE_IPKG=y" >> .config
   
    # 执行 defconfig，即使有警告也继续
    make defconfig || true
    popd >/dev/null
}
# =========================================================
# 3. 编译执行函数
# =========================================================
build_packages() {
    local BUILD_DIR="$1"
    local FMT="$2"
    local ARTIFACT_DIR="artifacts/$FMT"
    pushd "$BUILD_DIR" >/dev/null
    echo "--> 正在构建工具链 (RK3588)..."
    make tools/install -j$(nproc)
    make toolchain/install -j$(nproc)
   
    # 新添加：修复内核配置无效问题
    echo "--> 准备内核配置以修复无效配置错误..."
    make target/linux/clean V=s
    make target/linux/prepare V=s
   
    # 针对 Golang 项目的关键步骤：编译 host 端的 Go 环境
    echo "--> 准备 Go 编译环境..."
    make package/feeds/packages/golang/host/compile -j$(nproc)
    echo "--> 正在编译目标包..."
    make "package/custom/iptest/compile" -j$(nproc) V=s || true
    make "package/custom/luci-app-iptest/compile" -j$(nproc) V=s || true
    # 归档产物
    mkdir -p "../$ARTIFACT_DIR"
    find bin/packages -type f -name "*.$FMT" -exec cp {} "../$ARTIFACT_DIR/" \;
    popd >/dev/null
}
# =========================================================
# 4. 主流程逻辑
# =========================================================
mkdir -p artifacts
FORMAT_ARG="${1:-all}"
PLATFORM_ARG="${2:-1}" # 默认 1 指向 RK3588
# 确定要跑的平台
if [ "$PLATFORM_ARG" = "all" ]; then
    COMPILE_PLATFORMS=("${PLATFORMS[@]}")
else
    COMPILE_PLATFORMS=("${PLATFORMS[$PLATFORM_ARG]}")
fi
for PLATFORM_CFG in "${COMPILE_PLATFORMS[@]}"; do
    TARGET_NAME=$(echo "$PLATFORM_CFG" | tr '/' '_')
    [[ "$FORMAT_ARG" == "all" ]] && FMTS=("ipk" "apk") || FMTS=("$FORMAT_ARG")
    for FMT in "${FMTS[@]}"; do
        BUILD_DIR="build-${TARGET_NAME}-${FMT}"
        LOG_FILE="artifacts/log-${TARGET_NAME}-${FMT}.log"
       
        echo -e "\n[$(date +%T)] 任务启动: ${BUILD_DIR}"
        ( prepare_build_dir "$BUILD_DIR" "$PLATFORM_CFG" "$FMT" && \
          build_packages "$BUILD_DIR" "$FMT" ) 2>&1 | tee "$LOG_FILE"
    done
done
