FROM golang:1.24-alpine AS builder

WORKDIR /nfpcf

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /nfpcf

COPY --from=builder /nfpcf/bin/nfpcf /nfpcf/
COPY --from=builder /nfpcf/config/ /nfpcf/config/

EXPOSE 8000

CMD ["/nfpcf/nfpcf", "-c", "/nfpcf/config/nfpcfcfg.yaml"]
