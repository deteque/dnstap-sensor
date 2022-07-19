FROM debian:bullseye-slim
LABEL maintainer="Deteque Support <support@deteque.com>"
ENV GOLANG_VERSION "1.18.4"
ENV BUILD_DATE "2022-07-18"

WORKDIR /tmp
RUN mkdir /root/dnstap-sensor \
	&& mkdir /etc/dnstap \
	&& apt-get clean \
	&& apt-get update \
	&& apt-get -y dist-upgrade \
	&& apt-get install --no-install-recommends --no-install-suggests -y \
		apt-utils \
		build-essential \
		ca-certificates \
		dh-autoreconf \
		ethstats \
		libcap-dev \
		libcurl4-openssl-dev \
		libevent-dev \
		libpcap-dev \
		libssl-dev \
		net-tools \
		pkg-config \
		procps \
		sipcalc \
		sysstat \
		vim \
		wget 

COPY src/ /root/dnstap-sensor
 
WORKDIR /usr/local
RUN wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz \
	&& tar zxvf go${GOLANG_VERSION}.linux-amd64.tar.gz \
 	&& rm go${GOLANG_VERSION}.linux-amd64.tar.gz

WORKDIR /root/dnstap-sensor
RUN ./build.sh 

CMD ["/root/dnstap-sensor/dnstap-sensor"]
