#!/bin/sh
set -eu

: "${PHONE_NUMBER:?set PHONE_NUMBER in E.164 format}"

curl --fail-with-body -X POST http://localhost:8080/signup/code \
  -H 'Content-Type: application/json' \
  -d "{\"phone\":\"${PHONE_NUMBER}\",\"locale\":\"en\"}"

printf '\nEnter the code: '
read -r otp_code

curl --fail-with-body -X POST http://localhost:8080/signup/verify \
  -H 'Content-Type: application/json' \
  -d "{\"phone\":\"${PHONE_NUMBER}\",\"code\":\"${otp_code}\"}"
printf '\n'
