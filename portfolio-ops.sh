#!/bin/bash

# Portfolio Operations Tool - Bulk operations across all projects
# Usage: ./portfolio-ops.sh [operation] [pattern] [options]

PORTFOLIO_ROOT="/Users/bird/sources/standalone-projects"
OPERATION="$1"
PATTERN="$2"
shift 2

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Portfolio Operations Tool${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Helper function to check if project matches pattern
project_matches() {
    local project_dir="$1"
    local pattern="$2"

    if [ -z "$pattern" ]; then
        return 0  # No pattern means all projects
    fi

    case "$pattern" in
        "ai"|"ml")
            [[ "$project_dir" =~ (ai|ml|model|neural|gemma) ]]
            ;;
        "go"|"golang")
            [[ "$project_dir" =~ (go|golang) ]] || [ -f "$project_dir/go.mod" ]
            ;;
        "client"|"demo")
            [[ "$project_dir" =~ (client|demo|customer) ]]
            ;;
        "recent")
            # Projects modified in last 7 days
            local modified=$(find "$project_dir" -type f -mtime -7 2>/dev/null | head -1)
            [ -n "$modified" ]
            ;;
        "active")
            # Has git activity or recent changes
            [ -d "$project_dir/.git" ] && [ "$(git -C "$project_dir" log --since="1 week ago" --oneline | wc -l)" -gt 0 ]
            ;;
        *)
            # Default: pattern matching on directory name
            [[ "$project_dir" =~ $pattern ]]
            ;;
    esac
}

# Helper function to get project status
get_project_status() {
    local project_dir="$1"

    if [ ! -d "$project_dir" ]; then
        echo -e "${RED}❌ Missing${NC}"
        return
    fi

    local status=""
    local color=$GREEN

    # Check if it's a git repo
    if [ -d "$project_dir/.git" ]; then
        local uncommitted=$(git -C "$project_dir" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
        if [ "$uncommitted" -gt 0 ]; then
            status="📝 $uncommitted changes"
            color=$YELLOW
        else
            status="✅ Clean"
        fi

        # Check for recent commits
        local recent_commits=$(git -C "$project_dir" log --since="1 week ago" --oneline 2>/dev/null | wc -l | tr -d ' ')
        if [ "$recent_commits" -gt 0 ]; then
            status="$status 🔄 $recent_commits commits"
        fi
    else
        status="📁 No Git"
        color=$BLUE
    fi

    # Check for common files
    if [ -f "$project_dir/package.json" ]; then
        status="$status 🟦 Node.js"
    elif [ -f "$project_dir/go.mod" ]; then
        status="$status 🟩 Go"
    elif [ -f "$project_dir/requirements.txt" ] || [ -f "$project_dir/pyproject.toml" ]; then
        status="$status 🐍 Python"
    fi

    echo -e "${color}$status${NC}"
}

# Main operations
case "$OPERATION" in
    "list")
        echo -e "${GREEN}📊 Listing projects matching pattern: '$PATTERN'${NC}"
        echo ""

        count=0
        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [[ ! "$(basename "$dir")" =~ ^\. ]]; then
                if project_matches "$dir" "$PATTERN"; then
                    project_name=$(basename "$dir")
                    status=$(get_project_status "$dir")
                    printf "%-30s %s\n" "$project_name" "$status"
                    ((count++))
                fi
            fi
        done

        echo ""
        echo -e "${GREEN}Found $count matching projects${NC}"
        ;;

    "test")
        echo -e "${GREEN}🧪 Running tests across matching projects...${NC}"
        echo ""

        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && project_matches "$dir" "$PATTERN"; then
                project_name=$(basename "$dir")
                echo -e "${YELLOW}Testing $project_name...${NC}"

                (
                    cd "$dir"

                    # Node.js tests
                    if [ -f "package.json" ] && grep -q '"test"' package.json; then
                        npm test 2>/dev/null && echo -e "${GREEN}  ✓ Tests passed${NC}" || echo -e "${RED}  ✗ Tests failed${NC}"
                    # Go tests
                    elif [ -f "go.mod" ]; then
                        go test ./... 2>/dev/null && echo -e "${GREEN}  ✓ Tests passed${NC}" || echo -e "${RED}  ✗ Tests failed${NC}"
                    # Python tests
                    elif [ -f "pytest.ini" ] || [ -f "setup.py" ]; then
                        pytest 2>/dev/null && echo -e "${GREEN}  ✓ Tests passed${NC}" || echo -e "${RED}  ✗ Tests failed${NC}"
                    else
                        echo -e "${BLUE}  ℹ️  No tests configured${NC}"
                    fi
                )

                echo ""
            fi
        done
        ;;

    "status")
        echo -e "${GREEN}📈 Portfolio status overview${NC}"
        echo ""

        total_dirs=0
        git_repos=0
        active_repos=0
        node_projects=0
        go_projects=0
        python_projects=0

        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [[ ! "$(basename "$dir")" =~ ^\. ]]; then
                ((total_dirs++))

                if [ -d "$dir/.git" ]; then
                    ((git_repos++))
                    recent_commits=$(git -C "$dir" log --since="1 week ago" --oneline 2>/dev/null | wc -l | tr -d ' ')
                    if [ "$recent_commits" -gt 0 ]; then
                        ((active_repos++))
                    fi
                fi

                if [ -f "$dir/package.json" ]; then
                    ((node_projects++))
                elif [ -f "$dir/go.mod" ]; then
                    ((go_projects++))
                elif [ -f "$dir/requirements.txt" ] || [ -f "$dir/pyproject.toml" ]; then
                    ((python_projects++))
                fi
            fi
        done

        echo -e "Total Projects:        $total_dirs"
        echo -e "Git Repositories:      $git_repos"
        echo -e "Active (last week):    $active_repos"
        echo -e "Node.js Projects:      $node_projects"
        echo -e "Go Projects:          $go_projects"
        echo -e "Python Projects:       $python_projects"
        echo ""
        echo -e "${YELLOW}Most recently modified:${NC}"
        find "$PORTFOLIO_ROOT" -maxdepth 1 -type d -not -path "*/.*" -exec stat -f "%m %N" {} \; 2>/dev/null | sort -rn | head -6 | while read timestamp dir; do
            project_name=$(basename "$dir")
            if [ "$project_name" != "standalone-projects" ]; then
                date_formatted=$(date -r "$timestamp" "+%Y-%m-%d %H:%M")
                printf "  %-30s %s\n" "$project_name" "$date_formatted"
            fi
        done
        ;;

    "scan")
        echo -e "${GREEN}🔍 Scanning for common patterns...${NC}"
        echo ""

        echo -e "${YELLOW}Client-facing projects:${NC}"
        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [[ "$(basename "$dir")" =~ (client|customer|demo) ]]; then
                echo "  • $(basename "$dir")"
            fi
        done

        echo ""
        echo -e "${YELLOW}AI/ML projects:${NC}"
        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [[ "$(basename "$dir")" =~ (ai|ml|model|neural|gemma) ]]; then
                echo "  • $(basename "$dir")"
            fi
        done

        echo ""
        echo -e "${YELLOW}Service/feature extractions:${NC}"
        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [[ "$(basename "$dir")" =~ (service|extract|micro) ]]; then
                echo "  • $(basename "$dir")"
            fi
        done

        echo ""
        echo -e "${YELLOW}Projects with Docker:${NC}"
        docker_count=0
        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [ -f "$dir/Dockerfile" ] || [ -f "$dir/docker-compose.yml" ]; then
                echo "  • $(basename "$dir")"
                ((docker_count++))
            fi
        done
        if [ "$docker_count" -eq 0 ]; then
            echo "  None found"
        fi
        ;;

    "git-status")
        echo -e "${GREEN}📊 Git status across all projects${NC}"
        echo ""

        for dir in "$PORTFOLIO_ROOT"/*; do
            if [ -d "$dir" ] && [ -d "$dir/.git" ]; then
                project_name=$(basename "$dir")
                uncommitted=$(git -C "$dir" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
                branch=$(git -C "$dir" rev-parse --abbrev-ref HEAD 2>/dev/null)

                if [ "$uncommitted" -gt 0 ]; then
                    echo -e "${YELLOW}📝 $project_name${NC} ($branch) - $uncommitted uncommitted files"
                else
                    echo -e "${GREEN}✅ $project_name${NC} ($branch) - Clean"
                fi
            fi
        done
        ;;

    *)
        echo -e "${RED}❌ Unknown operation: $OPERATION${NC}"
        echo ""
        echo "Usage: $0 <operation> [pattern]"
        echo ""
        echo "Operations:"
        echo "  list [pattern]     - List projects matching pattern"
        echo "  test [pattern]     - Run tests across matching projects"
        echo "  status             - Show portfolio overview"
        echo "  scan               - Scan for common project patterns"
        echo "  git-status         - Show git status across all projects"
        echo ""
        echo "Patterns:"
        echo "  ai, ml             - AI/ML projects"
        echo "  go, golang         - Go projects"
        echo "  client, demo       - Client-facing projects"
        echo "  recent             - Recently modified projects"
        echo "  active             - Projects with recent git activity"
        echo "  [custom]           - Custom regex pattern on project names"
        exit 1
        ;;
esac