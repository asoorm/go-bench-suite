FROM golang:1.11-alpine as build

ARG VERSION

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache git && \
    go get -u github.com/asoorm/go-bench-suite && \
    cd /go/src/github.com/asoorm/go-bench-suite && git checkout --force $VERSION && \
    go install -a -ldflags="-s -w" .

FROM alpine:3.8

ENV HOST=0.0.0.0
ENV PORT=8081

RUN apk --no-cache add ca-certificates
RUN adduser -D -g bench bench
USER bench

WORKDIR /opt/bench
COPY --from=build /go/bin/go-bench-suite /opt/bench/go-bench-suite
USER bench

#EXPOSE $PORT

CMD ["./go-bench-suite"]
