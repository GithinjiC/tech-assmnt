#!/bin/bash

# Usage: chmod u+x restart_laravel.sh && ./restart_laravel.sh <service_name>

# macOS only. 
# The top command might not return expected output.

if [ -z "$1" ]; then
    echo "Usage: ./restart_laravel.sh <service_name>"
    exit

INTERVAL=30
LOG_FILE="laravel-monitor.log"

log_message() {
    local MESSAGE="$1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $MESSAGE" | tee -a "$LOG_FILE"
}

get_cpu_usage() {
    top_output=$(top -l 1 | grep "CPU usage")

    log_message "CPU Usage line: $top_output"

    cpu_idle=$(echo "$top_output" | awk -F "idle" '{print $1}' | awk '{print $NF}' | sed 's/%//')

    log_message "CPU idle percentage: $cpu_idle"

    if [[ -z "$cpu_idle" ]] || ! [[ "$cpu_idle" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
        log_message "Failed to get CPU idle percentage"
        exit 1
    fi

    cpu_usage=$(echo "scale=2; 100 - $cpu_idle" | bc)

    echo "$cpu_usage"
}

check_and_restart_service() {
    local cpu_usage=$1
    local threshold=80
    local service_name=$2

    if (( $(echo "$cpu_usage > $threshold" | bc -l) )); then
        log_message "CPU Usage ($cpu_usage%) is above $threshold%. Restarting $service_name"
        brew services restart $service_name
        log_message "$service_name restarted"
    else
        log_message "CPU usage ($cpu_usage%) is within acceptable limits"
    fi 
}

cpu_usage=$(get_cpu_usage)
check_and_restart_service "$cpu_usage" 