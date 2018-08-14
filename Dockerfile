FROM golang:1.8-alpine
RUN apk update
RUN apk add git --no-cache
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/America/Denver /etc/localtime
ADD root /var/spool/cron/crontabs/root
RUN mkdir -p /go/src/oct-redis-preprovision
ADD oct-redis-preprovision.go  /go/src/oct-redis-preprovision/oct-redis-preprovision.go
ADD create.sql /go/src/oct-redis-preprovision/create.sql
ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh
CMD ["crond", "-f"]
