# SPDX-FileCopyrightText: 2026 Vitaly Chekushkin
#
# SPDX-License-Identifier: AGPL-3.0-only

FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/betula ./cmd/betula

FROM alpine
RUN mkdir -p /data

ENV PORT=1738

COPY --from=builder /build/bin/betula /usr/local/bin/betula

EXPOSE 1738

ENTRYPOINT ["/usr/local/bin/betula", "/data/db.betula"]