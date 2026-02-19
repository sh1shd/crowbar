#!/bin/sh
# This script downloads Xray into a cache-aware directory so Docker BuildKit
# can reuse artifacts between builds. It prefers an explicit $CACHE_DIR,
# then /cache/xray (provided by a mount), then falls back to build/bin.

CACHE_DIR=${CACHE_DIR:-}
if [ -z "$CACHE_DIR" ] && [ -d /cache/xray ]; then
    CACHE_DIR=/cache/xray
fi
if [ -z "$CACHE_DIR" ]; then
    CACHE_DIR=build/bin
fi

case $1 in
    amd64)
        ARCH="64"
        FNAME="amd64"
        ;;
    armv8 | arm64 | aarch64)
        ARCH="arm64-v8a"
        FNAME="arm64"
        ;;
    *)
        ARCH="64"
        FNAME="amd64"
        ;;
esac

mkdir -p "$CACHE_DIR"
cd "$CACHE_DIR" || exit 1

if [ -x "xray-linux-${FNAME}" ]; then
    echo "xray-linux-${FNAME} already present in $CACHE_DIR, skipping download"
else
    echo "Downloading Xray for ${ARCH} -> ${FNAME} into $CACHE_DIR"
    curl -sfLRO "https://github.com/XTLS/Xray-core/releases/download/v26.2.6/Xray-linux-${ARCH}.zip"
    unzip "Xray-linux-${ARCH}.zip"
    rm -f "Xray-linux-${ARCH}.zip" geoip.dat geosite.dat
    mv xray "xray-linux-${FNAME}"
    curl -sfLRO https://cdn.jsdelivr.net/gh/Loyalsoldier/v2ray-rules-dat@release/geoip.dat
    curl -sfLRO https://cdn.jsdelivr.net/gh/Loyalsoldier/v2ray-rules-dat@release/geosite.dat
fi

cd - >/dev/null || true