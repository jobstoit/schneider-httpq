## NOTE this Dockerfile is used in combination with goreleaser
FROM alpine
ENTRYPOINT ["/usr/bin/httpq"]
COPY httpq /usr/bin/httpq
