#!/bin/sh

kill `ps -ef |grep madnet|grep [v]alidator3|awk '{print $2}'`

sleep 1

./madnet --config ./assets/config/validator3.toml validator > validator3-cont.log 2>&1 &

wait
