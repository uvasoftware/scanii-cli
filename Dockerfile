FROM scratch
COPY sc /sc
ENTRYPOINT ["/sc"]
