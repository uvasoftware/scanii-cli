FROM alpine:3
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
COPY sc /sc
ENTRYPOINT ["/sc"]
