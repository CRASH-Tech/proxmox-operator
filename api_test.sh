#!/bin/bash

API_URL=https://10.1.24.10:8006/api2/json
API_TOKEN=root@pam!test=9723a534-12eb-4956-8bdc-726f936d9f77
QUERY=$1

curl --insecure -H "Authorization: PVEAPIToken=${API_TOKEN}" ${API_URL}/${QUERY}