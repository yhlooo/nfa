FROM --platform=${TARGETPLATFORM} busybox:latest
COPY nfa /usr/bin/nfa
ENTRYPOINT ["/usr/bin/nfa"]
