FROM centos:7
USER root

MAINTAINER vichoudh@redhat.com

RUN yum install -y net-tools make which rsync lshw docker-client openssh-clients libcurl.i686

RUN \
  mkdir -p /goroot && \
  curl https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz | tar xvzf - -C /goroot --strip-components=1
# Set environment variables.
ENV GOROOT /goroot
ENV GOPATH /gopath
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

# Define working directory.
WORKDIR /gopath

COPY . $GOPATH/src/kubernetes/
RUN cd  $GOPATH/src/kubernetes \
#&& git apply patch4containerized.patch\
&& make WHAT=device_plugins/sfc_nic
RUN cp $GOPATH/src/kubernetes/_output/bin/sfc_nic /usr/bin/sfc-nic-plugin \
&& cp $GOPATH/src/kubernetes/device_plugins/sfc_nic/*.sh /usr/bin

ENTRYPOINT ["/usr/sbin/init"]
