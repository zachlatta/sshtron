ARG BASE_IMAGE=golang:latest
FROM $BASE_IMAGE as builder

WORKDIR $GOPATH/
RUN git clone https://github.com/zachlatta/sshtron.git
RUN ls /go/
WORKDIR $GOPATH/sshtron
RUN ls 

#ADD . .
# CGO_ENABLED=0 is here to fix this issue:
# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN git init
RUN go mod init github.com/zachlatta/sshtron

RUN go get && CGO_ENABLED=0 go build -o /usr/bin/sshtron .

FROM alpine:latest

COPY --from=builder /usr/bin/sshtron /usr/bin/
RUN apk add --update openssh-client && \
    ssh-keygen -t rsa -N "" -f id_rsa
ENTRYPOINT sshtron
