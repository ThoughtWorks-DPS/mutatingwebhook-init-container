#FROM gcr.io/distroless/base:nonroot
FROM busybox:glibc
WORKDIR /
COPY ./certificate-init-container ./certificate-init-container
ENTRYPOINT ["./certificate-init-container"]