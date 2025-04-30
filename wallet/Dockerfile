FROM golang:1.24.1-alpine AS builder

WORKDIR /usr/local/src

RUN  apk --no-cache add bash git make gcc gettext musl-dev sqlite

#dependencies
COPY ["./go.mod", "./go.sum", "./"]
RUN go mod download

#build
COPY . .
COPY internal/config/dev.yaml ./bin/internal/config/
RUN go build -o ./bin/cmd/app cmd/main.go

FROM alpine AS runner

COPY --from=builder --chown=nobody:nobody /usr/local/src/bin ./

CMD ["/cmd/app"]