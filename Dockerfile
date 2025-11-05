# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24-alpine
FROM golang:${GO_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.CurrentCommit=${COMMIT}" \
    -o /out/subcheck .

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S subcheck \
    && adduser -S -G subcheck subcheck

WORKDIR /app

COPY --from=builder /out/subcheck /app/subcheck
COPY --from=builder /src/config /app/config
COPY --from=builder /src/assets /app/assets

RUN chown -R subcheck:subcheck /app

USER subcheck

EXPOSE 14567
ENTRYPOINT ["/app/subcheck"]
CMD ["-f", "./config/config.yaml", "--port", "14567"]
