#!/usr/bin/env bash

curl -o -  -i -H "Content-Type: multipart/form-data" -F "haystackFile=@images/haystack.png" -F "needleFile=@images/needle1.png" -X POST http://localhost:8081/requests
