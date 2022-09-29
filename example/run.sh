#!/bin/bash
sudo sysctl -w net.core.rmem_max=2500000
#./example -v -qlog -isDropFirstInitialWithRetryToken -tcp -bind 0.0.0.0:6121
./example -v -qlog -isSendCCWhenServerReceiveInitialWithToken -tcp -bind 0.0.0.0:6121
