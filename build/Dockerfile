FROM registry.access.redhat.com/ubi7/ubi-minimal:7.7-98

ARG VCS_REF
ARG VCS_URL
ARG IMAGE_NAME
ARG IMAGE_DESCRIPTION
ARG ARCH_TYPE

COPY licenses /licenses

# http://label-schema.org/rc1/
LABEL org.label-schema.vendor="IBM" \
      org.label-schema.name="$IMAGE_NAME" \
      org.label-schema.description="$IMAGE_DESCRIPTION" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url=$VCS_URL \
      org.label-schema.license="Licensed Materials - Property of IBM" \
      org.label-schema.schema-version="1.0" \
      vendor="IBM" \
      name="$IMAGE_NAME" \
      summary="$IMAGE_DESCRIPTION" \
      description="$IMAGE_DESCRIPTION" \
      version="1.0" \
      release="$VCS_REF"

ENV OPERATOR=/usr/local/bin/ibm-klusterlet-operator \
    USER_UID=10001 \
    USER_NAME=ibm-klusterlet-operator

# install kubectl 
COPY build/kubectl-linux-${ARCH_TYPE} /usr/local/bin/kubectl

RUN chmod +x /usr/local/bin/kubectl

# install operator binary
COPY build/_output/bin/ibm-klusterlet-operator ${OPERATOR}

COPY build/bin /usr/local/bin

COPY component-operator/ /opt/component-operator/

RUN  /usr/local/bin/user_setup

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]

