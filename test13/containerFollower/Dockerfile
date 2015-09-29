FROM golang:onbuild
MAINTAINER Hylson Netto <hvescovi@gmail.com>
RUN mkdir /usr/mysoftware/
COPY follower.go /usr/mysoftware/
RUN go build -o /usr/mysoftware/follower /usr/mysoftware/follower.go
EXPOSE 8092
CMD /usr/mysoftware/follower
