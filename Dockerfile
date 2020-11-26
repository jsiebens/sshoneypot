FROM golang:1.13-alpine as build

RUN apk add upx

WORKDIR /go/src/github.com/jsiebens/sshoneypot

COPY .git               .git
COPY go.mod             .
COPY go.sum             .
COPY main.go            .
COPY pkg                pkg

RUN env ${OPTS} CGO_ENABLED=0 go build -ldflags "-s -w" -a -installsuffix cgo -o /usr/bin/sshoney \
    && addgroup -S app \
    && adduser -S -g app app

RUN upx -q -9 /usr/bin/sshoney

FROM scratch

COPY --from=build /etc/passwd /etc/group /etc/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/bin/sshoney /usr/bin/

USER app
EXPOSE 2222

ENTRYPOINT ["/usr/bin/sshoney"]