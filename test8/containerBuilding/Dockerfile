FROM golang:onbuild

MAINTAINER Hylson Netto <hvescovi@gmail.com>

RUN mkdir /usr/mysoftware/

COPY server.go /usr/mysoftware/

RUN go build -o /usr/mysoftware/server /usr/mysoftware/server.go

EXPOSE 8090

CMD /usr/mysoftware/server
