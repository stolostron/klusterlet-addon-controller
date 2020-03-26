FROM registry.access.redhat.com/ubi8/ubi-minimal:8.1-398

ARG ARCH_TYPE
ARG IMAGE_DESCRIPTION
ARG IMAGE_NAME
ARG IMAGE_VERSION
ARG VCS_REF
ARG VCS_URL

ENV OPERATOR=/usr/local/bin/endpoint-operator \
    USER_UID=10001 \
    USER_NAME=endpoint-operator

COPY licenses /licenses
COPY deploy/crds deploy/crds
COPY build/bin /usr/local/bin
COPY build/kubectl-linux-${ARCH_TYPE} /usr/local/bin/kubectl
COPY build/_output/bin/endpoint-operator ${OPERATOR}

# http://label-schema.org/rc1/
LABEL org.label-schema.vendor="IBM" \
      org.label-schema.name="$IMAGE_NAME" \
      org.label-schema.description="$IMAGE_DESCRIPTION" \
      org.label-schema.vcs-ref="$VCS_REF" \
      org.label-schema.vcs-url="$VCS_URL" \
      org.label-schema.license="Licensed Materials - Property of IBM" \
      org.label-schema.schema-version="1.0" \
      vendor="IBM" \
      name="$IMAGE_NAME" \
      summary="$IMAGE_DESCRIPTION" \
      description="$IMAGE_DESCRIPTION" \
      version="$IMAGE_VERSION" \
      release="$VCS_REF"

RUN chmod +x /usr/local/bin/kubectl
RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
