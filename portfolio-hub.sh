#!/bin/bash

# Portfolio Hub - Intelligent command discovery and execution for multi-project management
# Inspired by /command-hub AI routing, but optimized for portfolio operations

PORTFOLIO_ROOT="/Users/bird/sources/standalone-projects"
OPERATION="$1"
shift

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# AI Router Function - Inspired by command-hub
route_command() {
    local description="$*"
    local keywords=$(extract_keywords "$description")

    echo -e "${BLUE}🤖 Analyzing your request...${NC}"

    # High-confidence routing
    if [[ "$keywords" == *"status"* || "$keywords" == *"overview"* || "$keywords" == *"health"* ]]; then
        suggest_command "status" "Get portfolio and session overview" 95

    elif [[ "$keywords" == *"test"* || "$keywords" == *"quality"* || "$keywords" == *"check"* ]]; then
        suggest_command "test" "Run tests across projects" 90

    elif [[ "$keywords" == *"optimize"* || "$keywords" == *"improve"* || "$keywords" == *"fix"* ]]; then
        suggest_command "optimize" "Workspace optimization and suggestions" 92

    elif [[ "$keywords" == *"git"* || "$keywords" == *"commit"* || "$keywords" == *"changes"* ]]; then
        suggest_command "git-status" "Git status across all projects" 88

    elif [[ "$keywords" == *"create"* || "$keywords" == *"new"* || "$keywords" == *"start"* ]]; then
        suggest_command "create" "Create new project structure" 85

    elif [[ "$keywords" == *"suggest"* || "$keywords" == *"idea"* || "$keywords" == *"recommend"* ]]; then
        suggest_command "suggest" "Get intelligent project suggestions" 93

    elif [[ "$keywords" == *"client"* || "$keywords" == *"demo"* || "$keywords" == *"showcase"* ]]; then
        suggest_command "scan" "Scan for client-facing projects" 87

    elif [[ "$keywords" == *"ai"* || "$keywords" == *"ml"* || "$keywords" == *"model"* ]]; then
        suggest_command "list ai" "List all AI/ML projects" 89

    elif [[ "$keywords" == *"session"* || "$keywords" == *"manage"* || "$keywords" == *"organize"* ]]; then
        suggest_command "session-status" "Session management overview" 91

    else
        # Fuzzy matching and general suggestions
        fuzzy_route "$description"
    fi
}

extract_keywords() {
    local description="$*"
    echo "$description" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9 ]/ /g' | tr -s ' ' '\n' | grep -v '^$' | grep -E '^.{3,}$' | tr '\n' ' '
}

suggest_command() {
    local command="$1"
    local reasoning="$2"
    local confidence="$3"

    echo -e "${CYAN}🎯 Recommended Command:${NC}"
    echo -e "   ${GREEN}.//portfolio-hub.sh $command${NC}"

    local confidence_color=$GREEN
    if [ "$confidence" -lt 70 ]; then
        confidence_color=$YELLOW
    elif [ "$confidence" -lt 50 ]; then
        confidence_color=$RED
    fi

    echo -e "${confidence_color}   Confidence: ${confidence}%${NC}"
    echo -e "${BLUE}   Reasoning: $reasoning${NC}"

    # Show immediate execution option
    echo ""
    echo -e "${YELLOW}🚀 Execute now?${NC}"
    echo -e "   ${CYAN}.//portfolio-hub.sh $command${NC}"

    # Show alternatives if confidence is lower
    if [ "$confidence" -lt 80 ]; then
        echo ""
        echo -e "${BLUE}🔄 Alternative Commands:${NC}"
        echo -e "   • ./portfolio-hub.sh status    (${BLUE}Portfolio overview${NC})"
        echo -e "   • ./portfolio-hub.sh optimize   (${BLUE}Workspace optimization${NC})"
        echo -e "   • ./portfolio-hub.sh scan      (${BLUE}Pattern detection${NC})"
    fi
}

fuzzy_route() {
    local description="$*"
    local keywords=$(extract_keywords "$description")

    # Check for common patterns
    if [[ "$keywords" == *"help"* || "$keywords" == *"what"* || "$keywords" == *"available"* ]]; then
        show_help
        return
    fi

    # Default suggestions
    echo -e "${YELLOW}💡 General Recommendations:${NC}"
    echo ""
    echo -e "${CYAN}1. Portfolio Overview${NC}"
    echo -e "   ./portfolio-hub.sh status"
    echo -e "   ${BLUE}See all projects, git status, and recent activity${NC}"
    echo ""
    echo -e "${CYAN}2. Workspace Optimization${NC}"
    echo -e "   ./portfolio-hub.sh optimize"
    echo -e "   ${BLUE}Get intelligent suggestions for your workspace${NC}"
    echo ""
    echo -e "${CYAN}3. Smart Project Search${NC}"
    echo -e "   ./portfolio-hub.sh list [pattern]"
    echo -e "   ${BLUE}Find projects matching your interests${NC}"
    echo ""
    echo -e "${PURPLE}💭 Try being more specific:${NC}"
    echo -e "   \"I need to test all my AI projects\""
    echo -e "   \"Show me projects with uncommitted changes\""
    echo -e "   \"Optimize my workspace organization\""
}

show_help() {
    echo -e "${BLUE}🚀 Portfolio Hub - Intelligent Multi-Project Management${NC}"
    echo -e "${BLUE}=========================================================${NC}"
    echo ""
    echo -e "${CYAN}🎯 Smart Command Discovery:${NC}"
    echo -e "   ${GREEN}./portfolio-hub.sh \"your natural language request\"${NC}"
    echo -e "   Example: ./portfolio-hub.sh \"I need to test all my AI projects\""
    echo ""
    echo -e "${CYAN}📊 Direct Commands:${NC}"
    echo -e "   ${YELLOW}status${NC}                 - Portfolio and session overview"
    echo -e "   ${YELLOW}optimize${NC}               - Workspace optimization suggestions"
    echo -e "   ${YELLOW}git-status${NC}             - Git status across all projects"
    echo -e "   ${YELLOW}test [pattern]${NC}         - Run tests across projects"
    echo -e "   ${YELLOW}list [pattern]${NC}         - List projects matching pattern"
    echo -e "   ${YELLOW}scan${NC}                   - Scan for project patterns"
    echo -e "   ${YELLOW}session-status${NC}         - Session management overview"
    echo -e "   ${YELLOW}suggest \"idea\"${NC}        - Get project suggestions"
    echo -e "   ${YELLOW}create \"project\"${NC}      - Create new project"
    echo ""
    echo -e "${CYAN}🔍 Pattern Examples:${NC}"
    echo -e "   ${GREEN}./portfolio-hub.sh \"check the health of my projects\"${NC}"
    echo -e "   ${GREEN}./portfolio-hub.sh \"find client projects\"${NC}"
    echo -e "   ${GREEN}./portfolio-hub.sh \"run all tests\"${NC}"
    echo -e "   ${GREEN}./portfolio-hub.sh \"what needs to be committed\"${NC}"
    echo ""
    echo -e "${CYAN}🚀 Quick Start:${NC}"
    echo -e "   1. ${GREEN}./portfolio-hub.sh status${NC}                    # See your portfolio"
    echo -e "   2. ${GREEN}./portfolio-hub.sh optimize${NC}                  # Get optimization tips"
    echo -e "   3. ${GREEN}./portfolio-hub.sh \"your specific need\"${NC}     # AI-powered routing"
}

# Execute portfolio operations
execute_portfolio_ops() {
    local operation="$1"
    shift

    # Check if portfolio-ops.sh exists
    if [ ! -f "$PORTFOLIO_ROOT/portfolio-ops.sh" ]; then
        echo -e "${RED}❌ portfolio-ops.sh not found${NC}"
        return 1
    fi

    echo -e "${BLUE}🔧 Executing: ./portfolio-ops.sh $operation $@${NC}"
    echo ""
    "$PORTFOLIO_ROOT/portfolio-ops.sh" "$operation" "$@"
}

# Execute session orchestrator
execute_session_orchestrator() {
    local operation="$1"
    shift

    # Check if session-orchestrator.py exists
    if [ ! -f "$PORTFOLIO_ROOT/session-orchestrator.py" ]; then
        echo -e "${RED}❌ session-orchestrator.py not found${NC}"
        return 1
    fi

    echo -e "${BLUE}🎛️ Executing: python3 session-orchestrator.py $operation $@${NC}"
    echo ""
    python3 "$PORTFOLIO_ROOT/session-orchestrator.py" "$operation" "$@"
}

# Main command router
case "$OPERATION" in
    "status")
        echo -e "${BLUE}📊 Portfolio Status Overview${NC}"
        echo -e "${BLUE}============================${NC}"
        echo ""

        echo -e "${CYAN}🔍 Portfolio Health:${NC}"
        execute_portfolio_ops status
        echo ""

        echo -e "${CYAN}🎛️ Session Management:${NC}"
        execute_session_orchestrator status
        echo ""

        echo -e "${CYAN}🔥 Recent Activity:${NC}"
        execute_portfolio_ops list recent
        ;;

    "optimize")
        echo -e "${PURPLE}🧠 Workspace Intelligence Analysis${NC}"
        echo -e "${PURPLE}==================================${NC}"
        echo ""

        echo -e "${CYAN}📈 Portfolio Optimization:${NC}"
        execute_session_orchestrator optimize
        echo ""

        echo -e "${CYAN}🔍 Pattern Detection:${NC}"
        execute_portfolio_ops scan
        echo ""

        echo -e "${CYAN}💡 Smart Suggestions:${NC}"
        echo -e "   ${YELLOW}High session count?${NC} Use session orchestrator for allocation"
        echo -e "   ${YELLOW}Many uncommitted changes?${NC} Use './portfolio-hub.sh git-status'"
        echo -e "   ${YELLOW}Need project focus?${NC} Use './portfolio-hub.sh suggest \"your idea\"'"
        ;;

    "git-status")
        echo -e "${GREEN}📊 Git Status Across Portfolio${NC}"
        echo -e "${GREEN}===============================${NC}"
        execute_portfolio_ops git-status
        ;;

    "test")
        local pattern="$1"
        if [ -z "$pattern" ]; then
            echo -e "${YELLOW}🧪 Testing Active Projects${NC}"
            execute_portfolio_ops test active
        else
            echo -e "${YELLOW}🧪 Testing Projects: $pattern${NC}"
            execute_portfolio_ops test "$pattern"
        fi
        ;;

    "list")
        local pattern="$1"
        echo -e "${BLUE}📋 Project List${NC}"
        if [ -z "$pattern" ]; then
            execute_portfolio_ops list
        else
            execute_portfolio_ops list "$pattern"
        fi
        ;;

    "scan")
        echo -e "${PURPLE}🔍 Portfolio Pattern Scan${NC}"
        echo -e "${PURPLE}============================${NC}"
        execute_portfolio_ops scan
        ;;

    "session-status")
        echo -e "${CYAN}🎛️ Session Management Status${NC}"
        echo -e "${CYAN}=============================${NC}"
        execute_session_orchestrator status
        ;;

    "suggest")
        local idea="$*"
        if [ -z "$idea" ]; then
            echo -e "${RED}❌ Please provide an idea: ./portfolio-hub.sh suggest \"your idea\"${NC}"
            exit 1
        fi
        echo -e "${PURPLE}💭 Intelligent Project Suggestion${NC}"
        echo -e "${PURPLE}================================${NC}"
        execute_session_orchestrator suggest --idea "$idea"
        ;;

    "create")
        local project_name="$1"
        local project_type="${2:-general}"
        if [ -z "$project_name" ]; then
            echo -e "${RED}❌ Please provide project name: ./portfolio-hub.sh create \"project-name\"${NC}"
            exit 1
        fi
        echo -e "${GREEN}🚀 Creating Optimized Project${NC}"
        echo -e "${GREEN}============================${NC}"
        execute_session_orchestrator create --project "$project_name" --type "$project_type"
        ;;

    "help"|"-h"|"--help"|"")
        show_help
        ;;

    *)
        # AI Router for natural language requests
        if [ -z "$OPERATION" ]; then
            show_help
        else
            route_command "$OPERATION" "$@"
        fi
        ;;
esac