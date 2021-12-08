# dnstap-sensor
<p>DNSTAP-SENSOR is a Golang program that is used to collect passive dns data from a recursive name server and submit it to Deteque's DNSTAP collectors.</p>

# Overview
<p>Bind, Unbound and PowerDNS have native support for capturing DNS traffic using DNSTAP.  The Deteque DNSTAP-SENSOR is designed to accept DNSTAP from the nameserver and forward it securely to Deteque's collectors.</p>

<p>The DNSTAP-SENSOR application is a Go program that creates a socket that the nameserver connects to and accepts the dns log data.  As such, the DNSTAP-SENSOR should be started before the recursive nameserver daemon starts.  Additionally, the nameserver must be configured to connect to that socket.  While the socket can be mounted anywhere, we recommend that the socket be located in the same directory that DNSTAP-SENSOR's configuration file resides, which is "/etc/dnstap".</p>

# Documentation

Full installation and configuration settings are available at:<br>
https://deteque.com/dnstap-sensor/

# Downloading DNSTAP-SENSOR

Pre-compiled binaries are avaialble from the above URL.  We currently provide binaries for the following platforms:

- Linux amd64    Debian/Ubuntu/Redhat/Fedora/Amazon Linux
- Linux arm64.   Raspberry Pi OS
- Linux arm32.   Raspberry Pi OS
- Solaris amd64  Solaris
- FreeBSD amd64  FreeBSD
- OpenBSD amd64  OpenBSD
- NetBSD amd64   NetBSD

The source code and sample configuration file are is available three ways - as an https transfer, via Github or as a prebuilt Docker image:
- Web: https://deteque.com/dnstap-sensor/dnstap-sensor.tar.gz
- Docker: docker pull deteque/dnstap-sensor
- Git: git clone https://github.com/deteque/dnstap-sensor.git

# Installation Overview
Begin by creating a directory that will be used to store the DNSTAP-SENSOR configuration file and socket:
mkdir -p /etc/dnstap
