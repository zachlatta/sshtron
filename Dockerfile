FROM golang:latest

WORKDIR $GOPATH/src/github.com/zachlatta/sshtron

RUN apt-get update && apt-get install openssh-client && \
	ssh-keygen -t rsa -N "" -f id_rsa

ADD . .
RUN go get && go install

ENTRYPOINT sshtron
