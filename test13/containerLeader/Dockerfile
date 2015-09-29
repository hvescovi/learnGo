FROM golang:onbuild
MAINTAINER Hylson Netto <hvescovi@gmail.com>
RUN mkdir /usr/mysoftware/
COPY leader.go /usr/mysoftware/
RUN go build -o /usr/mysoftware/leader /usr/mysoftware/leader.go
EXPOSE 8091
CMD /usr/mysoftware/leader
