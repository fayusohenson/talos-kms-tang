FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags='-w -s' \
    -o /build/talos-kms-tang \
    ./cmd/server

FROM scratch

COPY --from=builder /build/talos-kms-tang /talos-kms-tang

EXPOSE 4050

USER 1001:1001

ENTRYPOINT ["/talos-kms-tang"]
