#!/bin/sh
#

curl -X GET http://yugo.moot.fr:8887/int.pem > int.pem 
curl -X GET http://yugo.moot.fr:8887/ca.pem > ca.pem 