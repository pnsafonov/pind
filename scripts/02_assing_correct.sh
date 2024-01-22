#!/usr/bin/env bash

for pid in $(pgrep -f /usr/bin/qemu-system-x86_64); do
    echo "pid=$pid"
    for tid in $(ls /proc/$pid/task); do
        echo "  tid=$tid"
        taskset -c -p $tid 
    done
    printf "\n"
done

