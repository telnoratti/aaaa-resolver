#!/bin/bash

go run server.go &

ret=$(dig @localhost -p 8053 +short AAAA 2001-db8--1.ipv6-literal)
if [[ ${ret} != "2001:db8::1" ]];
then
    echo "Failed the smell test"
fi

ret=$(dig @localhost -p 8053 +short AAAA 2001-db8--1.ipv6-literal-nozone)
if [[ ${ret} != "" ]];
then
    echo "Responding to extra zones"
fi

ret=$(dig @localhost -p 8053 +short AAAA 2001-db8--1.subdomain.ipv6-literal)
if [[ ${ret} != "" ]];
then
    echo "Responding to subdomains"
fi

ret=$(dig @localhost -p 8053 +short AAAA 192-168-0-1.ipv6-literal)
if [[ ${ret} != "" ]];
then
    echo "Responding to IPv4"
fi

ret=$(dig @localhost -p 8053 +short AAAA 192.168.0.1.ipv6-literal)
if [[ ${ret} != "" ]];
then
    echo "Responding to IPv4"
fi

ret=$(dig @localhost -p 8053 +short AAAA 2001-db8--.ipv6-literal)
if [[ ${ret} != "2001:db8::" ]];
then
    echo "Not responding to all v6"
fi

ret=$(dig @localhost -p 8053 +short AAAA 2001-db8---.ipv6-literal)
if [[ ${ret} != "" ]];
then
    echo "Responding to malformed IPv6"
fi

### cli testing
#kill 1
#sleep 2
#
#go run server.go --zone foo &
#
#ret=$(dig @localhost -p 8053 +short AAAA 2001-db8--.foo)
#if [[ ${ret} != "2001:db8::" ]];
#then
#    echo "Can't change zone"
#fi
#
#kill %1
#sleep 2
#
#go run server.go --port 8054 &
#
#ret=$(dig @localhost -p 8054 +short AAAA 2001-db8--.ipv6-literal)
#if [[ ${ret} != "2001:db8::" ]];
#then
#    echo "Can't change port"
#fi
