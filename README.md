# 🚀 Production Momentum Monitoring System

**Real-time development velocity tracking with iTerm2 integration for multi-session workflow optimization.**

## 🎯 Problem Solved

Tired of losing production momentum when juggling 10+ terminal sessions? This system prevents momentum slowdowns, manages multi-session chaos, and provides instant visibility across your entire development portfolio.

## 📊 What's Included

### **Core Components**
- **📈 momentum-monitor.sh** - Real-time momentum tracking with visual power bar
- **🤖 portfolio-hub.sh** - AI-powered command routing and portfolio optimization
- **⚡ portfolio-ops.sh** - Bulk operations across all your projects
- **🖥️ iterm2-momentum-integration.sh** - Complete iTerm2 integration system

### **Key Features**
- ✅ **Real-time momentum monitoring** (0-100% visual power bar)
- ✅ **Multi-session activity tracking** and optimization
- ✅ **AI-powered command routing** with confidence scoring
- ✅ **iTerm2 status bar integration** with live updates
- ✅ **Visual alerts** for momentum slowdowns
- ✅ **Automatic activity logging** and system performance monitoring

## 🚀 Quick Start

### **1. Clone and Install**
```bash
git clone https://github.com/birddigital/momentum-monitoring-system.git
cd momentum-monitoring-system

# Make all scripts executable
chmod +x *.sh
```

### **2. iTerm2 Integration (2 minutes)**

**Status Bar Setup:**
1. **iTerm2 → Preferences → Profiles → Session → Configure Status Bar**
2. **Add Component → Shell Command**
3. **Command:** `/path/to/momentum-monitor.sh status | grep 'Current Momentum' | cut -d: -f2 | tr -d ' %'`
4. **Update:** 30 seconds
5. **Colors:** Green ≥85, Yellow ≥70, Red <70

**Visual Triggers:**
1. **iTerm2 → Preferences → Profiles → Advanced → Triggers**
2. **Run:** `./iterm2-momentum-integration.sh install`

### **3. Test Your Setup**
```bash
# Quick momentum check
./momentum-monitor.sh status

# Full dashboard
./momentum-monitor.sh dashboard

# AI-powered portfolio hub
./portfolio-hub.sh "check the health of my projects"
```

## 🎮 Usage Examples

### **Natural Language Commands**
```bash
# AI routes to optimal command automatically
./portfolio-hub.sh "I need to test all my AI projects"
./portfolio-hub.sh "what needs to be committed"
./portfolio-hub.sh "optimize my workspace"
```

### **Momentum Monitoring**
```bash
# Continuous monitoring (recommended for dedicated terminal)
./momentum-monitor.sh monitor

# Quick check
./momentum-monitor.sh status

# Log your activity
./momentum-monitor.sh activity
```

### **Portfolio Operations**
```bash
# Overview of all projects
./portfolio-ops.sh status

# Test specific project types
./portfolio-ops.sh test ai

# Git status across portfolio
./portfolio-ops.sh git-status
```

## 📊 What You'll See

### **Status Bar Integration**
```
⚡ 95% [█████████░]    📝 5,873 changes    💻 12 sessions    🧠 66% memory
```

### **Visual Alerts**
- 🟢 **Green**: Momentum ≥ 85% (Optimal production)
- 🟡 **Yellow**: Momentum 70-84% (Good but could improve)
- 🔴 **Red**: Momentum < 70% (Alert! Take action)

### **Terminal Feedback**
- **Red flash**: Momentum alert detected
- **Yellow flash**: Memory warning
- **Blue flash**: Activity logged
- **Purple flash**: Tests passed

## 🎯 Benefits

### **Immediate**
- ✅ **Real-time momentum tracking** in your status bar
- ✅ **Visual alerts** for momentum slowdowns
- ✅ **One-command portfolio overview** across all projects
- ✅ **Automatic activity tracking**

### **Long-term**
- 🚀 **Never lose momentum** without noticing
- 📈 **Quantify your productivity patterns**
- 🎯 **Optimize your multi-session workflow**
- 💪 **Maintain peak production velocity**

## 🔧 Advanced Features

### **Custom Prompts**
Add to your `.zshrc`:
```bash
source "/path/to/.iterm2-momentum/momentum-shell.sh"
RPROMPT='$(momentum_prompt) $PROMPT'
```

### **Python API Integration**
```bash
# Install advanced monitoring
pip3 install iterm2
python3 /path/to/momentum-iterm2.py
```

### **Automatic Activity Tracking**
The system automatically logs activity when you run:
- `git` commands
- `npm` commands
- `go` commands
- `make` commands
- Any script starting with `./`

## 📋 System Requirements

- **macOS** with iTerm2 (recommended)
- **Bash** or **Zsh** shell
- **Git** (for portfolio tracking)
- **Optional**: Python 3 with iterm2 library

## 🛠️ Installation Scripts

### **Full Automated Setup**
```bash
# Install all components
./iterm2-momentum-integration.sh install

# Follow the on-screen instructions for iTerm2 configuration
```

### **Individual Components**
```bash
./iterm2-momentum-integration.sh shell      # Shell integration only
./iterm2-momentum-integration.sh python     # Python API only
./iterm2-momentum-integration.sh triggers   # Visual triggers only
```

## 📊 How It Works

### **Momentum Calculation**
The system calculates momentum based on:
- **Git activity** (uncommitted changes, recent commits)
- **System performance** (memory usage, CPU usage)
- **Session count** (active terminals)
- **Inactivity tracking** (time since last activity)
- **Taskflow integration** (if available)

### **Alert Thresholds**
- **🔴 Critical**: Momentum < 70% (Immediate action required)
- **🟡 Warning**: Memory > 80% or Momentum 70-84%
- **🟢 Optimal**: Momentum ≥ 85%

## 🤝 Contributing

This system addresses the **Multi-Session Fragmentation Syndrome (MSFS)** - the challenge of maintaining productivity when juggling multiple development sessions.

Feel free to:
- Submit issues for edge cases
- Suggest improvements to the momentum algorithm
- Share your iTerm2 configurations
- Report bugs with specific terminal environments

## 📄 License

MIT License - Feel free to use, modify, and distribute.

## 🌟 Credits

Built with ❤️ for high-velocity developers who refuse to let momentum slow down.

**Keep your production velocity high!** 🚀

---

*Generated with [Claude Code](https://claude.com/claude-code)*

Co-Authored-By: Claude <noreply@anthropic.com>