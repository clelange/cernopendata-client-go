FROM scratch
COPY cernopendata-client-go /
ENTRYPOINT ["/cernopendata-client-go"]