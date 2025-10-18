#!/bin/bash
#
# Repository Cleanup Script
# Removes backup files, archives old session reports, and maintains clean git status
#
# Usage:
#   ./scripts/cleanup-repository.sh --dry-run  # Preview changes
#   ./scripts/cleanup-repository.sh            # Execute cleanup
#

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

DRY_RUN=false
if [ "$1" = "--dry-run" ]; then
    DRY_RUN=true
    echo -e "${YELLOW}=== DRY RUN MODE ===${NC}"
    echo "No files will be modified"
    echo ""
fi

REPO_ROOT="/Users/anuoamdutta/Desktop/statuspage/Beakon"
cd "$REPO_ROOT"

echo -e "${GREEN}=== Beakon Repository Cleanup ===${NC}"
echo "Repository: $REPO_ROOT"
echo ""

# Counter for removed files
BACKUP_COUNT=0
SESSION_DOC_COUNT=0

# Step 1: Remove backup files
echo -e "${BLUE}Step 1: Removing backup files${NC}"
echo "Finding *.bak, *.backup.*, *.old files..."

BACKUP_FILES=$(find microservices -type f \( -name "*.bak" -o -name "*.backup.*" -o -name "*.old" -o -name "*~" \) 2>/dev/null || true)

if [ -n "$BACKUP_FILES" ]; then
    echo "$BACKUP_FILES" | while read -r file; do
        if [ "$DRY_RUN" = true ]; then
            echo -e "${YELLOW}  Would remove: $file${NC}"
        else
            rm -f "$file"
            echo -e "${GREEN}  ✓ Removed: $file${NC}"
        fi
        BACKUP_COUNT=$((BACKUP_COUNT + 1))
    done
    echo -e "${GREEN}Found $BACKUP_COUNT backup files${NC}"
else
    echo -e "${GREEN}✓ No backup files found${NC}"
fi

echo ""

# Step 2: List session-specific docs for review
echo -e "${BLUE}Step 2: Session-specific documents (review recommended)${NC}"

SESSION_DOCS=(
    "microservices/landing-page-service/IMPLEMENTATION_FIX_SUMMARY.md"
    "microservices/landing-page-service/IMPLEMENTATION_SUMMARY.md"
    "microservices/landing-page-service/LANDING_PAGE_ENHANCEMENT.md"
)

echo "The following session-specific docs should be reviewed:"
for doc in "${SESSION_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${YELLOW}  ⚠ $doc${NC}"
        SESSION_DOC_COUNT=$((SESSION_DOC_COUNT + 1))
    fi
done

if [ $SESSION_DOC_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ No session docs found${NC}"
else
    echo ""
    echo "To archive these files:"
    echo "  mkdir -p docs/archives/landing-page-service"
    echo "  mv microservices/landing-page-service/*_SUMMARY.md docs/archives/landing-page-service/"
fi

echo ""

# Step 3: Check git status for uncommitted deletions
echo -e "${BLUE}Step 3: Checking for uncommitted deletions${NC}"

DELETED_COUNT=$(git status --short | grep -c "^ D" || true)

if [ $DELETED_COUNT -gt 0 ]; then
    echo -e "${YELLOW}Found $DELETED_COUNT deleted files not yet committed:${NC}"
    git status --short | grep "^ D" | head -10
    if [ $DELETED_COUNT -gt 10 ]; then
        echo "  ... and $((DELETED_COUNT - 10)) more"
    fi
    echo ""
    echo "To commit these deletions:"
    echo "  git add -u"
    echo "  git commit -m 'chore: remove obsolete documentation and config files'"
else
    echo -e "${GREEN}✓ No uncommitted deletions${NC}"
fi

echo ""

# Step 4: Create/update .gitignore
echo -e "${BLUE}Step 4: Updating .gitignore${NC}"

GITIGNORE_ENTRIES=(
    "# Backup and temporary files"
    "*.bak"
    "*.backup"
    "*.backup.*"
    "*.old"
    "*~"
    "# OS files"
    ".DS_Store"
    "Thumbs.db"
)

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}Would add/update .gitignore entries${NC}"
else
    # Check if .gitignore exists
    if [ ! -f .gitignore ]; then
        echo "# Beakon .gitignore" > .gitignore
    fi

    # Add entries if they don't exist
    for entry in "${GITIGNORE_ENTRIES[@]}"; do
        if ! grep -q "$entry" .gitignore 2>/dev/null; then
            echo "$entry" >> .gitignore
            echo -e "${GREEN}  ✓ Added: $entry${NC}"
        fi
    done
fi

echo ""

# Step 5: Summary
echo -e "${GREEN}=== Cleanup Summary ===${NC}"
echo "Backup files found: $BACKUP_COUNT"
echo "Session docs for review: $SESSION_DOC_COUNT"
echo "Uncommitted deletions: $DELETED_COUNT"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}This was a DRY RUN - no files were modified${NC}"
    echo "Run without --dry-run to execute cleanup"
else
    echo -e "${GREEN}Cleanup complete!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review session-specific docs and archive if needed"
    echo "  2. Commit git deletions: git add -u && git commit -m 'chore: cleanup'"
    echo "  3. Review DOCUMENTATION_CLEANUP_REPORT.md for additional recommendations"
fi
