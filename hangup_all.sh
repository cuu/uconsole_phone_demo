#!/bin/sh

CUR=`mmcli -L | cut -d "/" -f 6 | cut -d " " -f 1`

mmcli -m $CUR --voice-list-calls

sudo mmcli -m $CUR --voice-hangup-all

