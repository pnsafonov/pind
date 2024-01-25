#!/usr/bin/env bash

# failed
#ls -1 /proc/$(pgrep -f deb-2)/task | xargs -I$ taskset -cp 0-1 $

# ok
for pid in $(pgrep -f deb-2); do
    echo "pid=$pid"
    for tid in $(ls /proc/$pid/task); do
        echo "  tid=$tid"
        taskset -cp 0-1 $tid
    done
    printf "\n"
done

