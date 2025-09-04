#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building static binary using Alpine Linux container${NC}"

cat > Dockerfile.static-build << 'EOF'
FROM golang:1.25-alpine

RUN apk add --no-cache \
    git \
    build-base \
    autoconf \
    automake \
    libtool \
    linux-headers \
    alsa-lib-dev

WORKDIR /tmp

# Build ALSA library statically
RUN git clone https://github.com/alsa-project/alsa-lib.git && \
    cd alsa-lib && \
    git checkout v1.2.10 && \
    libtoolize --force --copy --automake && \
    aclocal && \
    autoheader && \
    automake --foreign --copy --add-missing && \
    autoconf && \
    ./configure --prefix=/usr/local --enable-shared=no --enable-static=yes --disable-ucm && \
    make -j$(nproc) && \
    make install

WORKDIR /app
COPY . .

ENV PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH"
ENV CGO_CFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib"
RUN CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static -L/usr/local/lib"' -o 2n-static

RUN file 2n-static && (ldd 2n-static 2>/dev/null || echo "✓ Static binary!") && ls -lh 2n-static
EOF

echo -e "${YELLOW}Building...${NC}"
docker build -f Dockerfile.static-build -t 2n-static-builder .
docker run --rm -v "$(pwd):/output" 2n-static-builder cp /app/2n-static /output/
rm Dockerfile.static-build

echo -e "${GREEN}Success! Static binary: 2n-static${NC}"
file 2n-static
ldd 2n-static 2>/dev/null || echo "✓ Truly static!"