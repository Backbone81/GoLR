# This Dockerfile is used by GoReleaser to create a multi-arch container image for GoLR.
FROM scratch

ARG TARGETPLATFORM

COPY $TARGETPLATFORM/golr /golr

ENTRYPOINT ["/golr"]
