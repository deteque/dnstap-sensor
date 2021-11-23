#!/bin/sh
/usr/bin/docker run \
	--name dnstap-sensor \
	--rm \
	--detach \
	-v /etc/dnstap:/etc/dnstap \
	/usr/sbin/dnstap-sensor
