#!/bin/sh

VERSION=${VERSION:-0.0.12}

curl -L https://github.com/ffromani/numalign/releases/download/v${VERSION}/numalign-v${VERSION}-linux-amd64 -o numalign
chmod 0755 numalign
