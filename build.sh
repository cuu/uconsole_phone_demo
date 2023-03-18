#!/bin/sh

echo 'building uconsole_phone_demo'

go build -o uconsole_phone_demo demo.go modem.go 
