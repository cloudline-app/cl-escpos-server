#! /bin/sh

while true; do
  [ -e stopme ] && break
  ./redial-print-server
  sleep 1
done
