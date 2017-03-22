FROM golang:latest

ENV PROJECT_NAME sshtron
ENV PROJECT_PATH github.com/zachlatta/sshtron

ADD . $GOPATH/src/$PROJECT_PATH
WORKDIR $GOPATH/src/$PROJECT_PATH

RUN apt-get update && apt-get install openssh-client && \
	ssh-keygen -t rsa -N "" -f id_rsa && \
	go get && go install && \
	rm -rf /var/lib/apt/lists/*

ENTRYPOINT sshtron

