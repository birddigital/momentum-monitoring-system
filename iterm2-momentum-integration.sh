#!/bin/bash

# iTerm2 Momentum Integration - Production momentum monitoring in your terminal
# Multiple integration approaches for iTerm2

MOMENTUM_ROOT="/Users/bird/sources/standalone-projects"
ITERM2_INTEGRATION_DIR="$HOME/.iterm2-momentum"

# Colors for iTerm2 integration
GREEN='\\033[38;2;0;255;0m'
YELLOW='\\033[38;2;255;255;0m'
RED='\\033[38;2;255;0;0m'
BLUE='\\033[38;2;0;123;255m'
PURPLE='\\033[38;2;147;0;211m'
CYAN='\\033[38;2;0;255;255m'
NC='\\033[0m'

# Create integration directory
mkdir -p "$ITERM2_INTEGRATION_DIR"

# Get current momentum (simplified version for iTerm2)
get_momentum_quick() {
    local momentum_file="$HOME/.momentum/current_momentum"
    local momentum=100

    if [ -f "$momentum_file" ]; then
        momentum=$(cat "$momentum_file" 2>/dev/null || echo "100")
    fi

    # Calculate quick momentum
    local git_changes=$(find "$MOMENTUM_ROOT" -name ".git" -exec git -C {} status --porcelain \; 2>/dev/null | wc -l | tr -d ' ')
    local active_terminals=$(ps aux | grep -E "(iTerm|Terminal)" | grep -v grep | wc -l | tr -d ' ')
    local memory_usage=$(ps -o %mem -ax | awk '{sum+=$1} END {printf "%.0f", sum}')

    # Quick momentum calculation
    if [ "$git_changes" -gt 100 ]; then
        momentum=85
    elif [ "$git_changes" -gt 50 ]; then
        momentum=90
    elif [ "$git_changes" -gt 10 ]; then
        momentum=95
    fi

    if [ "$memory_usage" -gt 80 ]; then
        momentum=$((momentum - 20))
    elif [ "$memory_usage" -gt 60 ]; then
        momentum=$((momentum - 10))
    fi

    if [ "$active_terminals" -gt 15 ]; then
        momentum=$((momentum - 15))
    elif [ "$active_terminals" -gt 10 ]; then
        momentum=$((momentum - 5))
    fi

    [ "$momentum" -lt 0 ] && momentum=0
    [ "$momentum" -gt 100 ] && momentum=100

    echo "$momentum"
}

# Get momentum color for display
get_momentum_color() {
    local momentum="$1"
    if [ "$momentum" -ge 85 ]; then
        echo "$GREEN"
    elif [ "$momentum" -ge 70 ]; then
        echo "$YELLOW"
    else
        echo "$RED"
    fi
}

# Create iTerm2 shell integration
create_shell_integration() {
    local shell_rc="$1"

    cat > "$ITERM2_INTEGRATION_DIR/momentum-shell.sh" << 'EOF'
# iTerm2 Momentum Shell Integration
# Source this in your .zshrc or .bashrc

momentum_prompt() {
    local momentum_file="$HOME/.momentum/current_momentum"
    local momentum=100

    if [ -f "$momentum_file" ]; then
        momentum=$(cat "$momentum_file" 2>/dev/null || echo "100")
    fi

    # Calculate quick momentum
    local git_changes=$(find "$HOME/sources/standalone-projects" -name ".git" -exec git -C {} status --porcelain \; 2>/dev/null | wc -l | tr -d ' ')
    local memory_usage=$(ps -o %mem -ax | awk '{sum+=$1} END {printf "%.0f", sum}')
    local active_terminals=$(ps aux | grep -E "(iTerm|Terminal)" | grep -v grep | wc -l | tr -d ' ')

    # Quick momentum calculation
    if [ "$git_changes" -gt 100 ]; then
        momentum=85
    elif [ "$git_changes" -gt 50 ]; then
        momentum=90
    elif [ "$git_changes" -gt 10 ]; then
        momentum=95
    fi

    if [ "$memory_usage" -gt 80 ]; then
        momentum=$((momentum - 20))
    elif [ "$memory_usage" -gt 60 ]; then
        momentum=$((momentum - 10))
    fi

    if [ "$active_terminals" -gt 15 ]; then
        momentum=$((momentum - 15))
    elif [ "$active_terminals" -gt 10 ]; then
        momentum=$((momentum - 5))
    fi

    [ "$momentum" -lt 0 ] && momentum=0
    [ "$momentum" -gt 100 ] && momentum=100

    # Color based on momentum
    local color=""
    if [ "$momentum" -ge 85 ]; then
        color="%F{green}"
    elif [ "$momentum" -ge 70 ]; then
        color="%F{yellow}"
    else
        color="%F{red}"
    fi

    # Create momentum bar
    local bar=""
    local filled_length=$((momentum / 10))
    for ((i=0; i<10; i++)); do
        if [ "$i" -lt "$filled_length" ]; then
            bar+="█"
        else
            bar+="░"
        fi
    done

    echo "${color}[${bar}] ${momentum}%%f"
}

# Add to prompt (customize as needed)
# RPROMPT='$(momentum_prompt)'
# Or add to your existing prompt

# Quick momentum check
alias momentum-check='echo "Momentum: $(momentum_prompt)"'

# Update momentum activity
alias momentum-activity='date +%s > $HOME/.momentum/last_activity && echo "Activity logged"'

# Quick momentum commands
alias momentum-status='$HOME/sources/standalone-projects/momentum-monitor.sh status'
alias momentum-monitor='$HOME/sources/standalone-projects/momentum-monitor.sh monitor'

# Auto-update momentum on key commands
preexec() {
    # Update activity when running important commands
    if [[ "$1" =~ ^(git|npm|go|python|node|make|./) ]]; then
        date +%s > "$HOME/.momentum/last_activity"
    fi
}
EOF

    echo "✅ Shell integration created at $ITERM2_INTEGRATION_DIR/momentum-shell.sh"
}

# Create iTerm2 Python API integration
create_python_integration() {
    cat > "$ITERM2_INTEGRATION_DIR/momentum-iterm2.py" << 'EOF'
#!/usr/bin/env python3
"""
iTerm2 Momentum Monitor Integration
Real-time momentum monitoring in iTerm2 using Python API
"""

import asyncio
import iterm2
import subprocess
import time
import os
from datetime import datetime

async def update_momentum_bar(connection, app):
    """Update iTerm2 status bar with momentum"""
    try:
        # Get momentum from our monitor
        result = subprocess.run(
            ['/Users/bird/sources/standalone-projects/momentum-monitor.sh', 'status'],
            capture_output=True, text=True, timeout=5
        )

        if result.returncode == 0:
            # Parse momentum from output
            for line in result.stdout.split('\n'):
                if 'Current Momentum:' in line:
                    try:
                        momentum = int(line.split(':')[1].strip().rstrip('%'))
                        break
                    except:
                        momentum = 100
            else:
                momentum = 100
        else:
            momentum = 100

        # Determine color
        if momentum >= 85:
            color = "green"
            bar_color = "green"
        elif momentum >= 70:
            color = "yellow"
            bar_color = "yellow"
        else:
            color = "red"
            bar_color = "red"

        # Create momentum bar
        filled_length = momentum // 10
        bar = "█" * filled_length + "░" * (10 - filled_length)

        # Create status string
        status = f"⚡ {momentum}% [{bar}]"

        # Update status bar
        await app.async_set_variable("user.momentum_status", status)
        await app.async_set_variable("user.momentum_color", color)
        await app.async_set_variable("user.momentum_value", str(momentum))

        # Set window title if momentum is low
        if momentum < 70:
            await app.async_set_variable("user.title", f"🚨 LOW MOMENTUM: {momentum}%")

    except Exception as e:
        # Fallback on error
        await app.async_set_variable("user.momentum_status", "⚡ 100%")
        await app.async_set_variable("user.momentum_color", "green")
        await app.async_set_variable("user.momentum_value", "100")

async def monitor_momentum(connection):
    """Main monitoring loop"""
    app = await iterm2.async_get_app(connection)

    print("🚀 iTerm2 Momentum Monitor Started")
    print("Press Ctrl+C to stop")

    # Update immediately
    await update_momentum_bar(connection, app)

    # Update every 30 seconds
    while True:
        await asyncio.sleep(30)
        await update_momentum_bar(connection, app)

def main():
    iterm2.run_until_complete(monitor_momentum)

if __name__ == "__main__":
    main()
EOF

    chmod +x "$ITERM2_INTEGRATION_DIR/momentum-iterm2.py"
    echo "✅ Python integration created at $ITERM2_INTEGRATION_DIR/momentum-iterm2.py"
}

# Create iTerm2 triggers configuration
create_triggers_config() {
    cat > "$ITERM2_INTEGRATION_DIR/iterm2-triggers.txt" << 'EOF'
# iTerm2 Momentum Triggers Configuration
# Import these in iTerm2 Preferences > Profiles > Advanced > Triggers

# Low Momentum Alert (RED)
Regular Expression: 🚨 MOMENTUM ALERT
Action: Set Background Color
Color: Red (100,0,0) with 20% opacity

# High Memory Alert (YELLOW)
Regular Expression: 🚨 MEMORY ALERT
Action: Set Background Color
Color: Yellow (100,100,0) with 20% opacity

# Momentum Warning (YELLOW)
Regular Expression: Momentum at [0-6][0-9]%
Action: Set Background Color
Color: Yellow (100,100,0) with 15% opacity

# Good Momentum (GREEN)
Regular Expression: Momentum at [8-9][0-9]%
Action: Set Background Color
Color: Green (0,100,0) with 10% opacity

# Perfect Momentum (GREEN)
Regular Expression: Momentum at 100%
Action: Set Background Color
Color: Green (0,150,0) with 10% opacity

# Activity Updates (BLUE FLASH)
Regular Expression: ✅ Activity logged
Action: Set Background Color
Color: Blue (0,100,255) with 10% opacity

# Test Results (PURPLE)
Regular Expression: ✓ Tests passed
Action: Set Background Color
Color: Purple (150,0,255) with 10% opacity

# Test Failures (RED FLASH)
Regular Expression: ✗ Tests failed
Action: Set Background Color
Color: Red (255,0,0) with 30% opacity

# Git Status (CYAN)
Regular Expression: 📝 [0-9]+ uncommitted files
Action: Set Background Color
Color: Cyan (0,255,255) with 10% opacity
EOF

    echo "✅ Triggers configuration created at $ITERM2_INTEGRATION_DIR/iterm2-triggers.txt"
}

# Create status bar configuration
create_statusbar_config() {
    cat > "$ITERM2_INTEGRATION_DIR/statusbar-config.txt" << 'EOF'
# iTerm2 Status Bar Configuration for Momentum Monitoring
# Add these to your status bar in iTerm2 Preferences > Profiles > Session > Configure Status Bar

# Momentum Status (Custom Component)
Component Name: Momentum
Type: Interpolated String
Format: \(user.momentum_status\)
Action: None
Colors: Dynamic (use \(user.momentum_color\))

# Manual Entry (if Python API doesn't work)
Component Name: Momentum
Type: Shell Command
Command: /Users/bird/sources/standalone-projects/momentum-monitor.sh status | grep "Current Momentum" | cut -d: -f2 | tr -d ' %'
Update Interval: 30 seconds
Colors:
  - Green: >= 85
  - Yellow: >= 70
  - Red: < 70

# Git Changes Counter
Component Name: Git Changes
Type: Shell Command
Command: find /Users/bird/sources/standalone-projects -name ".git" -exec git -C {} status --porcelain \; 2>/dev/null | wc -l
Update Interval: 60 seconds

# Session Count
Component Name: Sessions
Type: Shell Command
Command: ps aux | grep -E "(iTerm|Terminal)" | grep -v grep | wc -l
Update Interval: 30 seconds

# Memory Usage
Component Name: Memory
Type: Shell Command
Command: ps -o %mem -ax | awk '{sum+=$1} END {printf "%.0f%%", sum}'
Update Interval: 30 seconds
EOF

    echo "✅ Status bar configuration created at $ITERM2_INTEGRATION_DIR/statusbar-config.txt"
}

# Create automatic installation script
create_installer() {
    cat > "$ITERM2_INTEGRATION_DIR/install.sh" << 'EOF'
#!/bin/bash

# iTerm2 Momentum Integration Installer
# Automated setup for all integration options

ITERM2_INTEGRATION_DIR="$HOME/.iterm2-momentum"
SHELL_RC="$HOME/.zshrc"

echo "🚀 Installing iTerm2 Momentum Integration..."
echo "=========================================="

# Check shell type
if [[ $SHELL == *"zsh"* ]]; then
    SHELL_RC="$HOME/.zshrc"
elif [[ $SHELL == *"bash"* ]]; then
    SHELL_RC="$HOME/.bashrc"
else
    echo "⚠️  Unsupported shell: $SHELL"
    echo "Manual integration required"
fi

echo "📁 Shell detected: $SHELL_RC"

# Install shell integration
echo ""
echo "1. Installing shell integration..."
if ! grep -q "momentum-shell.sh" "$SHELL_RC" 2>/dev/null; then
    echo "" >> "$SHELL_RC"
    echo "# iTerm2 Momentum Integration" >> "$SHELL_RC"
    echo "source \"$ITERM2_INTEGRATION_DIR/momentum-shell.sh\"" >> "$SHELL_RC"
    echo "✅ Shell integration added to $SHELL_RC"
else
    echo "✅ Shell integration already exists"
fi

# Install status bar components
echo ""
echo "2. Status Bar Configuration:"
echo "   Open iTerm2 Preferences > Profiles > Session > Configure Status Bar"
echo "   Add the following components:"
echo "   - Shell Command: momentum-monitor.sh status | grep 'Current Momentum' | cut -d: -f2 | tr -d ' %'"
echo "   - Shell Command: find ~/sources/standalone-projects -name .git -exec git -C {} status --porcelain \\; | wc -l"
echo "   - Shell Command: ps aux | grep -E '(iTerm|Terminal)' | grep -v grep | wc -l"

# Install triggers
echo ""
echo "3. Trigger Configuration:"
echo "   Open iTerm2 Preferences > Profiles > Advanced > Triggers"
echo "   Import the file: $ITERM2_INTEGRATION_DIR/iterm2-triggers.txt"

# Python API integration (optional)
echo ""
echo "4. Python API Integration (optional):"
echo "   Install with: pip3 install iterm2"
echo "   Run with: python3 $ITERM2_INTEGRATION_DIR/momentum-iterm2.py"

echo ""
echo "✅ Installation complete!"
echo ""
echo "🔄 Restart your shell or run: source $SHELL_RC"
echo ""
echo "🎯 Quick test commands:"
echo "   momentum-check    # Show current momentum"
echo "   momentum-status   # Full momentum status"
echo "   momentum-activity # Log activity"
EOF

    chmod +x "$ITERM2_INTEGRATION_DIR/install.sh"
    echo "✅ Installer created at $ITERM2_INTEGRATION_DIR/install.sh"
}

# Main installation function
install_all() {
    echo "🚀 Creating iTerm2 Momentum Integration..."
    echo "=========================================="

    create_shell_integration
    create_python_integration
    create_triggers_config
    create_statusbar_config
    create_installer

    echo ""
    echo "✅ All integration components created!"
    echo ""
    echo "📁 Installation directory: $ITERM2_INTEGRATION_DIR"
    echo ""
    echo "🔧 To install:"
    echo "   $ITERM2_INTEGRATION_DIR/install.sh"
    echo ""
    echo "📚 Components created:"
    echo "   • momentum-shell.sh      - Shell integration"
    echo "   • momentum-iterm2.py    - Python API integration"
    echo "   • iterm2-triggers.txt   - Visual triggers"
    echo "   • statusbar-config.txt  - Status bar components"
    echo "   • install.sh            - Automated installer"
}

# Show usage
show_usage() {
    echo "🚀 iTerm2 Momentum Integration"
    echo "==============================="
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  install    - Create all integration components"
    echo "  shell      - Create shell integration only"
    echo "  python     - Create Python API integration only"
    echo "  triggers   - Create triggers configuration only"
    echo "  statusbar  - Create status bar configuration only"
    echo "  help       - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 install     # Install everything"
    echo "  $0 shell       # Shell integration only"
    echo "  $0 python      # Python API only"
    echo ""
    echo "After installation:"
    echo "  1. Run: $ITERM2_INTEGRATION_DIR/install.sh"
    echo "  2. Restart shell or: source ~/.zshrc"
    echo "  3. Configure iTerm2 status bar and triggers"
}

# Main command routing
case "${1:-install}" in
    "install")
        install_all
        ;;
    "shell")
        create_shell_integration
        echo "✅ Shell integration created"
        ;;
    "python")
        create_python_integration
        echo "✅ Python integration created"
        ;;
    "triggers")
        create_triggers_config
        echo "✅ Triggers configuration created"
        ;;
    "statusbar")
        create_statusbar_config
        echo "✅ Status bar configuration created"
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo "❌ Unknown command: $1"
        show_usage
        exit 1
        ;;
esac