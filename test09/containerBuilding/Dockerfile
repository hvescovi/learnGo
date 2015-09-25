FROM golang:onbuild
MAINTAINER Hylson Netto <hvescovi@gmail.com>
RUN mkdir /usr/mysoftware/
COPY master.go /usr/mysoftware/
RUN go build -o /usr/mysoftware/master /usr/mysoftware/master.go
EXPOSE 8090
CMD /usr/mysoftware/master
