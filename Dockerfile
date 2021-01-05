FROM alpine:3.7
RUN adduser -D -u 10000 kubelog
RUN apk add --update ca-certificates
COPY kubelog /
RUN chown kubelog /kubelog
USER kubelog
ENTRYPOINT ["/kubelog"]
