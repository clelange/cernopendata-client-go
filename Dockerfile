FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY cernopendata-client-go /
RUN chmod +x ./cernopendata-client-go
ENTRYPOINT ["/cernopendata-client-go"]