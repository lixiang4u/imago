#!/bin/bash

git pull
go get

#CGO_ENABLED=1  CGO_CFLAGS="-std=gnu99"  PKG_CONFIG_PATH=/usr/local/vips-8.15.0/build/lib64/pkgconfig/ go build .
CGO_ENABLED=1  CGO_CFLAGS="-std=gnu99"  PKG_CONFIG_PATH=/usr/local/vips-8.15.0/build/lib64/pkgconfig/ go build imago.go
CGO_ENABLED=1  CGO_CFLAGS="-std=gnu99"  PKG_CONFIG_PATH=/usr/local/vips-8.15.0/build/lib64/pkgconfig/ go build api.go


