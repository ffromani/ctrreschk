FROM alpine:3.9
COPY _out /usr/local/bin
ENTRYPOINT ["/bin/sh"]
