#!/bin/bash

# Production Momentum Monitor - Real-time development velocity tracking
# Prevents momentum slowdown and maintains high-velocity development

MOMENTUM_ROOT="/Users/bird/sources/standalone-projects"
MOMENTUM_STATE_DIR="$HOME/.momentum"
ALERT_THRESHOLD=70    # Alert when momentum drops below 70%
MEMORY_THRESHOLD=80   # Alert when memory usage above 80%
INACTIVITY_MINUTES=5  # Alert after 5 minutes of inactivity

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Ensure momentum state directory exists
mkdir -p "$MOMENTUM_STATE_DIR"

# Initialize momentum tracking
init_momentum() {
    local timestamp=$(date +%s)
    echo "$timestamp" > "$MOMENTUM_STATE_DIR/last_activity"
    echo "100" > "$MOMENTUM_STATE_DIR/current_momentum"
    echo "$timestamp" > "$MOMENTUM_STATE_DIR/session_start"

    # Create momentum log
    echo "timestamp,momentum,git_changes,active_sessions,memory_usage" > "$MOMENTUM_STATE_DIR/momentum_log.csv"
}

# Get current system metrics
get_system_metrics() {
    local memory_usage=$(ps -o %mem -ax | awk '{sum+=$1} END {printf "%.0f", sum}')
    local cpu_usage=$(ps -o %cpu -ax | awk '{sum+=$1} END {printf "%.0f", sum}')
    local active_terminals=$(ps aux | grep -E "(iTerm|Terminal|tmux|screen)" | grep -v grep | wc -l | tr -d ' ')
    local vscode_processes=$(ps aux | grep "Code" | grep -v grep | wc -l | tr -d ' ')

    # Ensure we have valid numbers
    [ -z "$memory_usage" ] && memory_usage=0
    [ -z "$cpu_usage" ] && cpu_usage=0
    [ -z "$active_terminals" ] && active_terminals=0
    [ -z "$vscode_processes" ] && vscode_processes=0

    echo "$memory_usage,$cpu_usage,$active_terminals,$vscode_processes"
}

# Get development activity metrics
get_dev_metrics() {
    local git_changes=0
    local active_projects=0
    local recent_commits=0
    local current_time=$(date +%s)
    local five_minutes_ago=$((current_time - 300))

    # Count uncommitted changes across all projects
    for dir in "$MOMENTUM_ROOT"/*; do
        if [ -d "$dir" ] && [ -d "$dir/.git" ]; then
            local changes=$(git -C "$dir" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
            git_changes=$((git_changes + changes))

            if [ "$changes" -gt 0 ]; then
                active_projects=$((active_projects + 1))
            fi

            # Check for recent commits
            local latest_commit=$(git -C "$dir" log -1 --format="%ct" 2>/dev/null)
            if [ "$latest_commit" -gt "$five_minutes_ago" ]; then
                recent_commits=$((recent_commits + 1))
            fi
        fi
    done

    echo "$git_changes,$active_projects,$recent_commits"
}

# Check taskflow activity
check_taskflow() {
    local taskflow_dir="$MOMENTUM_ROOT/taskflow"
    if [ -d "$taskflow_dir" ]; then
        # Check for recent taskflow activity
        local recent_files=$(find "$taskflow_dir" -type f -mmin -5 2>/dev/null | wc -l | tr -d ' ')
        echo "$recent_files"
    else
        echo "0"
    fi
}

# Calculate momentum score
calculate_momentum() {
    local system_metrics_str=$(get_system_metrics)
    IFS=',' read -ra system_metrics <<< "$system_metrics_str"
    local memory_usage=${system_metrics[0]:-0}
    local cpu_usage=${system_metrics[1]:-0}
    local active_terminals=${system_metrics[2]:-0}
    local vscode_processes=${system_metrics[3]:-0}

    local dev_metrics_str=$(get_dev_metrics)
    IFS=',' read -ra dev_metrics <<< "$dev_metrics_str"
    local git_changes=${dev_metrics[0]:-0}
    local active_projects=${dev_metrics[1]:-0}
    local recent_commits=${dev_metrics[2]:-0}

    local taskflow_activity=$(check_taskflow)

    # Get last activity time
    local last_activity=$(cat "$MOMENTUM_STATE_DIR/last_activity" 2>/dev/null || echo "$(date +%s)")
    local current_time=$(date +%s)
    local minutes_since_activity=$(((current_time - last_activity) / 60))

    # Momentum calculation (0-100)
    local momentum=100

    # Deduct for high memory usage
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        momentum=$((momentum - (memory_usage - MEMORY_THRESHOLD)))
    fi

    # Deduct for inactivity
    if [ "$minutes_since_activity" -gt "$INACTIVITY_MINUTES" ]; then
        momentum=$((momentum - (minutes_since_activity - INACTIVITY_MINUTES) * 5))
    fi

    # Boost for active development
    if [ "$git_changes" -gt 0 ]; then
        momentum=$((momentum + git_changes / 2))
    fi

    if [ "$recent_commits" -gt 0 ]; then
        momentum=$((momentum + recent_commits * 10))
    fi

    if [ "$taskflow_activity" -gt 0 ]; then
        momentum=$((momentum + taskflow_activity * 5))
    fi

    # Boost for multiple active sessions (up to a point)
    if [ "$active_terminals" -gt 1 ] && [ "$active_terminals" -lt 10 ]; then
        momentum=$((momentum + active_terminals * 2))
    elif [ "$active_terminals" -ge 10 ]; then
        momentum=$((momentum - 10))  # Too many sessions = chaos
    fi

    # Ensure momentum stays within bounds
    if [ "$momentum" -gt 100 ]; then
        momentum=100
    elif [ "$momentum" -lt 0 ]; then
        momentum=0
    fi

    echo "$momentum,$memory_usage,$cpu_usage,$git_changes,$active_projects,$recent_commits,$taskflow_activity,$minutes_since_activity"
}

# Display momentum power bar
display_power_bar() {
    local momentum="$1"
    local momentum_int=${momentum%.*}  # Remove decimal if present

    # Determine color based on momentum level
    local bar_color=$GREEN
    if [ "$momentum_int" -lt "$ALERT_THRESHOLD" ]; then
        bar_color=$RED
    elif [ "$momentum_int" -lt 85 ]; then
        bar_color=$YELLOW
    fi

    # Create power bar
    local bar_length=50
    local filled_length=$((momentum_int * bar_length / 100))
    local empty_length=$((bar_length - filled_length))

    # Build bar
    local bar=""
    for ((i=0; i<filled_length; i++)); do
        bar+="█"
    done
    for ((i=0; i<empty_length; i++)); do
        bar+="░"
    done

    echo -e "${bar_color}[${bar}]${NC} ${momentum_int}%"
}

# Check for alerts and trigger notifications
check_alerts() {
    local momentum="$1"
    local memory_usage="$2"
    local minutes_since_activity="$8"
    local momentum_int=${momentum%.*}

    local alert_triggered=false

    # Momentum slowdown alert
    if [ "$momentum_int" -lt "$ALERT_THRESHOLD" ]; then
        echo -e "${RED}🚨 MOMENTUM ALERT: Production slowdown detected!${NC}"
        echo -e "${RED}   Momentum at ${momentum_int}% (threshold: ${ALERT_THRESHOLD}%)${NC}"

        # Check taskflow if available
        if [ -d "$MOMENTUM_ROOT/taskflow" ]; then
            echo -e "${YELLOW}   Pinging taskflow to check status...${NC}"
            (cd "$MOMENTUM_ROOT/taskflow" && ls -la 2>/dev/null) || echo -e "${RED}   ⚠️ Taskflow unreachable${NC}"
        fi

        alert_triggered=true
    fi

    # Memory usage alert
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        echo -e "${RED}🚨 MEMORY ALERT: System resources constrained!${NC}"
        echo -e "${RED}   Memory usage at ${memory_usage}% (threshold: ${MEMORY_THRESHOLD}%)${NC}"
        echo -e "${YELLOW}   💡 Consider closing unused applications or sessions${NC}"
        alert_triggered=true
    fi

    # Inactivity alert
    if [ "$minutes_since_activity" -gt "$INACTIVITY_MINUTES" ]; then
        echo -e "${YELLOW}⏰ INACTIVITY NOTICE: No development activity for ${minutes_since_activity} minutes${NC}"
        echo -e "${YELLOW}   💡 Stay active to maintain momentum${NC}"
    fi

    # System performance recommendations
    if [ "$alert_triggered" = true ]; then
        echo -e "${CYAN}📋 Quick Actions:${NC}"
        echo -e "   • Run: ./portfolio-hub.sh status (check project health)"
        echo -e "   • Run: ./portfolio-hub.sh git-status (review uncommitted changes)"
        echo -e "   • Run: ./portfolio-hub.sh optimize (workspace suggestions)"
        echo ""
    fi

    return $([ "$alert_triggered" = true ] && echo 1 || echo 0)
}

# Update activity timestamp
update_activity() {
    date +%s > "$MOMENTUM_STATE_DIR/last_activity"
}

# Log momentum data
log_momentum() {
    local timestamp=$(date +%s)
    local momentum_data="$1"
    local momentum=${momentum_data%%,*}

    # Append to CSV log
    echo "$timestamp,$momentum_data" >> "$MOMENTUM_STATE_DIR/momentum_log.csv"

    # Keep only last 1000 entries
    tail -n 1000 "$MOMENTUM_STATE_DIR/momentum_log.csv" > "$MOMENTUM_STATE_DIR/momentum_log.tmp"
    mv "$MOMENTUM_STATE_DIR/momentum_log.tmp" "$MOMENTUM_STATE_DIR/momentum_log.csv"
}

# Display momentum dashboard
display_dashboard() {
    clear
    echo -e "${BLUE}🚀 PRODUCTION MOMENTUM MONITOR${NC}"
    echo -e "${BLUE}=============================${NC}"
    echo ""

    # Calculate current momentum
    local momentum_data=$(calculate_momentum)
    IFS=',' read -r momentum memory_usage cpu_usage git_changes active_projects recent_commits taskflow_activity minutes_since_activity <<< "$momentum_data"

    # Power bar
    echo -e "${CYAN}📊 Momentum Power Bar:${NC}"
    display_power_bar "$momentum"
    echo ""

    # Metrics panel
    echo -e "${CYAN}📈 System Metrics:${NC}"
    echo -e "   Memory Usage: ${memory_usage}%"
    echo -e "   CPU Usage: ${cpu_usage}%"
    echo -e "   Active Terminals: $(ps aux | grep -E "(iTerm|Terminal)" | grep -v grep | wc -l | tr -d ' ')"
    echo -e "   VS Code Processes: $(ps aux | grep "Code" | grep -v grep | wc -l | tr -d ' ')"
    echo ""

    echo -e "${CYAN}💻 Development Activity:${NC}"
    echo -e "   Git Changes: $git_changes uncommitted files"
    echo -e "   Active Projects: $active_projects with changes"
    echo -e "   Recent Commits: $recent_commits (last 5 min)"
    echo -e "   Taskflow Activity: $taskflow_activity recent files"
    echo -e "   Last Activity: $minutes_since_activity minutes ago"
    echo ""

    # Session duration
    local session_start=$(cat "$MOMENTUM_STATE_DIR/session_start" 2>/dev/null || echo "$(date +%s)")
    local session_duration=$(((date +%s) - session_start) / 60)
    echo -e "${CYAN}⏱️ Session Duration: ${session_duration} minutes${NC}"
    echo ""

    # Check for alerts
    check_alerts "$momentum_data"

    # Recent momentum trend
    echo -e "${CYAN}📊 Recent Momentum Trend:${NC}"
    tail -n 10 "$MOMENTUM_STATE_DIR/momentum_log.csv" | tail -n 5 | while IFS=',' read -r timestamp momentum_val memory cpu git active recent taskflow inactive; do
        local time_str=$(date -r "$timestamp" "+%H:%M:%S")
        local momentum_int=${momentum_val%.*}
        local color=$GREEN
        [ "$momentum_int" -lt "$ALERT_THRESHOLD" ] && color=$RED
        printf "   ${time_str}: ${color}%3d%%${NC}\n" "$momentum_int"
    done
}

# Continuous monitoring mode
monitor_mode() {
    echo -e "${GREEN}🔍 Starting continuous momentum monitoring...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
    echo ""

    # Initialize if needed
    if [ ! -f "$MOMENTUM_STATE_DIR/last_activity" ]; then
        init_momentum
    fi

    while true; do
        display_dashboard
        update_activity

        # Calculate momentum for next loop
        local momentum_data=$(calculate_momentum)
        local momentum=${momentum_data%%,*}
        local momentum_int=${momentum%.*}

        # Determine sleep time based on momentum
        local sleep_time=30
        if [ "$momentum_int" -lt "$ALERT_THRESHOLD" ]; then
            sleep_time=15  # Check more frequently when momentum is low
        fi

        sleep "$sleep_time"
    done
}

# Quick status check
quick_status() {
    if [ ! -f "$MOMENTUM_STATE_DIR/last_activity" ]; then
        init_momentum
    fi

    local momentum_data=$(calculate_momentum)
    local momentum=${momentum_data%%,*}
    local momentum_int=${momentum%.*}

    echo -e "${CYAN}⚡ Current Momentum: ${momentum_int}%${NC}"
    display_power_bar "$momentum"

    if [ "$momentum_int" -lt "$ALERT_THRESHOLD" ]; then
        echo -e "${RED}🚨 Momentum below threshold! Consider action.${NC}"
    else
        echo -e "${GREEN}✅ Momentum looking good!${NC}"
    fi
}

# Main command routing
case "${1:-status}" in
    "init")
        init_momentum
        echo -e "${GREEN}✅ Momentum monitoring initialized${NC}"
        ;;

    "monitor"|"watch")
        monitor_mode
        ;;

    "status"|"quick")
        quick_status
        ;;

    "dashboard"|"full")
        if [ ! -f "$MOMENTUM_STATE_DIR/last_activity" ]; then
            init_momentum
        fi
        display_dashboard
        ;;

    "activity")
        update_activity
        echo -e "${GREEN}✅ Activity timestamp updated${NC}"
        ;;

    "reset")
        init_momentum
        echo -e "${YELLOW}🔄 Momentum tracking reset${NC}"
        ;;

    "history")
        if [ -f "$MOMENTUM_STATE_DIR/momentum_log.csv" ]; then
            echo -e "${CYAN}📊 Momentum History (last 20 entries):${NC}"
            tail -n 20 "$MOMENTUM_STATE_DIR/momentum_log.csv" | column -t -s','
        else
            echo -e "${YELLOW}No history available yet${NC}"
        fi
        ;;

    "help"|"-h"|"--help")
        echo -e "${BLUE}🚀 Production Momentum Monitor${NC}"
        echo -e "${BLUE}============================${NC}"
        echo ""
        echo -e "${CYAN}Commands:${NC}"
        echo -e "  ${GREEN}status${NC}        - Quick momentum check"
        echo -e "  ${GREEN}dashboard${NC}     - Full momentum dashboard"
        echo -e "  ${GREEN}monitor${NC}       - Continuous monitoring mode"
        echo -e "  ${GREEN}activity${NC}      - Update activity timestamp"
        echo -e "  ${GREEN}init${NC}          - Initialize tracking"
        echo -e "  ${GREEN}reset${NC}         - Reset momentum tracking"
        echo -e "  ${GREEN}history${NC}       - Show momentum history"
        echo ""
        echo -e "${CYAN}Examples:${NC}"
        echo -e "  ${YELLOW}./momentum-monitor.sh monitor${NC}     # Start live monitoring"
        echo -e "  ${YELLOW}./momentum-monitor.sh status${NC}      # Quick check"
        echo -e "  ${YELLOW}./momentum-monitor.sh activity${NC}    # Log activity"
        echo ""
        echo -e "${CYAN}Integration:${NC}"
        echo -e "  Add ${YELLOW}./momentum-monitor.sh activity${NC} to your git hooks"
        echo -e "  Run ${YELLOW}./momentum-monitor.sh monitor${NC} in a dedicated terminal"
        ;;

    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo -e "${YELLOW}Use: $0 help${NC} for usage information"
        exit 1
        ;;
esac