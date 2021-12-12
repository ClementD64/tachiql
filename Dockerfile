FROM golang:1.17-alpine as builder

ARG BUILD_PKG=./cmd/tachiql

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o tachiql ${BUILD_PKG}

FROM gcr.io/distroless/static

COPY --from=builder /app/tachiql /tachiql
ENTRYPOINT [ "/tachiql" ]