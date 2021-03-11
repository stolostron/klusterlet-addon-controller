# Copyright Contributors to the Open Cluster Management project

FROM openshift/origin-release:golang-1.13 as builder

ARG REMOTE_SOURCE
ENV REMOTE_SOURCE_DIR=/remote-source 

COPY . $REMOTE_SOURCE_DIR/app/
WORKDIR $REMOTE_SOURCE_DIR/app

RUN GOFLAGS="" go build ./cmd/manager

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
RUN microdnf update

ENV REMOTE_SOURCE_DIR=/remote-source 

ENV IMAGE_MANIFEST_PATH=/

ENV OPERATOR=/usr/local/bin/klusterlet-addon-controller \
    USER_UID=10001 \
    USER_NAME=klusterlet-addon-controller

COPY --from=builder $REMOTE_SOURCE_DIR/app/deploy/crds deploy/crds
COPY --from=builder $REMOTE_SOURCE_DIR/app/manager ${OPERATOR}
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin/entrypoint /usr/local/bin
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin/user_setup /usr/local/bin

LABEL org.label-schema.vendor="Red Hat" \
    org.label-schema.name="$IMAGE_NAME" \
    org.label-schema.description="$IMAGE_DESCRIPTION" \
    org.label-schema.vcs-ref="$VCS_REF" \
    org.label-schema.vcs-url="$VCS_URL" \
    org.label-schema.license="Red Hat Advanced Cluster Management for Kubernetes EULA" \
    org.label-schema.schema-version="1.0" \
    vendor="Red Hat" \
    name="$IMAGE_NAME" \
    summary="$IMAGE_DESCRIPTION" \
    description="$IMAGE_DESCRIPTION" \
    version="$IMAGE_VERSION" \
    release="$VCS_REF"

RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
