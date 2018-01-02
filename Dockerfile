FROM golang:latest as builder

WORKDIR $GOPATH/src/github.com/zachlatta/sshtron
ADD . .
RUN go get && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/bin/sshtron .

FROM alpine:latest

COPY --from=builder /usr/bin/sshtron /usr/bin/
RUN apk add --update --no-cache openssh-client && \
	ssh-keygen -t rsa -N "" -f id_rsa
ENTRYPOINT sshtron
