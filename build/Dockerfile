# Copyright Contributors to the Open Cluster Management project

FROM registry.ci.openshift.org/stolostron/builder:go1.22-linux AS builder

ARG REMOTE_SOURCE
ARG REMOTE_SOURCE_DIR

COPY $REMOTE_SOURCE $REMOTE_SOURCE_DIR/app/
WORKDIR $REMOTE_SOURCE_DIR/app

RUN GOFLAGS="" go build ./cmd/manager && \
    GOFLAGS="" go test \
    -covermode=atomic  \
    -coverpkg=github.com/stolostron/klusterlet-addon-controller/pkg/...  \
    -c -tags testrunmain \
    ./cmd/manager \
    -o manager-coverage

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

RUN microdnf update -y && \
    microdnf clean all

ARG REMOTE_SOURCE_DIR

ENV IMAGE_MANIFEST_PATH=/

ENV OPERATOR=/usr/local/bin/klusterlet-addon-controller \
    USER_UID=10001 \
    USER_NAME=klusterlet-addon-controller

COPY --from=builder $REMOTE_SOURCE_DIR/app/deploy/crds deploy/crds
COPY --from=builder $REMOTE_SOURCE_DIR/app/manager ${OPERATOR}
COPY --from=builder $REMOTE_SOURCE_DIR/app/manager-coverage ${OPERATOR}-coverage
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin /usr/local/bin
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/coverage-entrypoint-func.sh /usr/local/bin/coverage-entrypoint-func.sh

RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
