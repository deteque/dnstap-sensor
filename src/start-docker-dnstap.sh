#!/bin/sh
/usr/bin/docker run \
	--rm \
	--detach \
	--name dnstap-sensor \
	--volume /etc/dnstap:/etc/dnstap \
	deteque/dnstap-sensor
