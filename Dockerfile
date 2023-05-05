ARG BASE_IMAGE="fedora:37"

FROM ${BASE_IMAGE}

ARG OCM_CLI_VERSION="0.1.60"
ARG RHOC_CLI_VERSION="0.0.37"
ARG BACKPLANE_CLI_VERSION="0.1.2"

RUN dnf update -y && \
    dnf install -y procps \
    wget \
    golang \
    sudo \
    jq \
    python3 \
    python-pip \
    git \
    htop && \
    yum install -y net-tools \
    make && \
    pip install jinja2 && \
    curl -Lo /usr/bin/ocm https://github.com/openshift-online/ocm-cli/releases/download/v${OCM_CLI_VERSION}/ocm-linux-amd64 && \
    chmod +x /usr/bin/ocm && \
    wget https://github.com/openshift/backplane-cli/releases/download/v${BACKPLANE_CLI_VERSION}/ocm-backplane_${BACKPLANE_CLI_VERSION}_Linux_x86_64.tar.gz && \
    tar -xvf ocm-backplane_${BACKPLANE_CLI_VERSION}_Linux_x86_64.tar.gz && \
    mv $PWD/ocm-backplane /usr/bin/ocm-backplane && \
    chmod +x /usr/bin/ocm-backplane && \
    curl -Lo /usr/bin/rhoc https://github.com/bf2fc6cc711aee1a0c2a/cos-tools/releases/download/v${RHOC_CLI_VERSION}/rhoc_${RHOC_CLI_VERSION}_linux_amd64.tar.gz && \
    chmod +x /usr/bin/rhoc && \
    wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/stable/openshift-client-linux.tar.gz && \
    tar -xvf openshift-client-linux.tar.gz && \
    mv $PWD/oc /usr/bin/oc && \
    mv $PWD/kubectl /usr/bin/kubectl

RUN mkdir /ocm-workspace
WORKDIR /ocm-workspace
ADD ./workspace .
