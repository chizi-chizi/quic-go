#!/bin/bash
sysctl -w net.core.rmem_max=2500000

#rm *.qlog keylog*
#./client -insecure -repeatCnt 100 -onlySendInitial  -q  https://127.0.0.1:6121
#./client -insecure -repeatCnt 100000 -onlySendInitial  -q  https://www.google.com/ https://www.cloudflare.com https://www.facebook.com/ 
#./client -insecure -repeatCnt 100 -onlySendInitial  -q  https://www.google.com/ https://www.cloudflare.com https://www.facebook.com/ 
#./client -insecure -repeatCnt 10000 -onlySendInitial  -q  https://127.0.0.1:6121
./client -insecure  -q  https://127.0.0.1:8443
