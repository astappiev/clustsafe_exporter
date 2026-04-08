ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/clustsafe_exporter   /bin/clustsafe_exporter

USER nobody
EXPOSE     9879
ENTRYPOINT [ "/bin/clustsafe_exporter" ]
