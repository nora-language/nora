#!/usr/bin/env bash

# Premium colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# ASCII Art Banner
print_banner() {
    echo -e "${PURPLE}${BOLD}"
    echo " ███▄    █   ▒█████   ██▀███   ▄▄▄      "
    echo " ██ ▀█   █  ▒██▒  ██▒▓██ ▒ ██▒▒████▄    "
    echo "▓██  ▀█ ██▒▒██░  ██▒▓██ ░▄█ ▒▒██  ▀█▄  "
    echo "▓██▒  ▐▌██▒▒██   ██░▒██▀▀█▄  ░██▄▄▄▄██ "
    echo "▒██░   ▓██░░ ████▓▒░░██▓ ▒██▒ ▓█   ▓██▒"
    echo "░ ▒░   ▒ ▒ ░ ▒░▒░▒░ ░ ▒▓ ░▒▓░ ▒▒   ▓▒█░"
    echo "░ ░░   ░ ▒░  ░ ▒ ▒░   ░▒ ░ ▒░  ▒   ▒▒ ░"
    echo "   ░   ░ ░ ░ ░ ░ ▒    ░░   ░   ░   ▒   "
    echo "         ░     ░ ░     ░           ░  ░"
    echo -e "${NC}"
}

clear
print_banner
echo -e "${CYAN}${BOLD}=== NORA LANGUAGE SYSTEM INSTALLER ===${NC}\n"

# Step 1: Detect Platform
echo -e "${CYAN}[1/4] Detecting OS and Architecture...${NC}"
OS_NAME="$(uname -s)"
ARCH_NAME="$(uname -m)"
echo -e "  - Target OS: ${GREEN}${OS_NAME}${NC}"
echo -e "  - Architecture: ${GREEN}${ARCH_NAME}${NC}"

# Step 2: Establish Install Paths
NORA_DIR="$HOME/.nora"
BIN_DIR="$NORA_DIR/bin"
STD_DIR="$NORA_DIR/std"

echo -e "\n${CYAN}[2/4] Setting up installation directories...${NC}"
echo -e "  - Destination: ${YELLOW}${NORA_DIR}${NC}"
mkdir -p "$BIN_DIR"
mkdir -p "$STD_DIR"
echo -e "  ${GREEN}✓${NC} Directories successfully prepared."

# Step 3: Compile / Install compiler and runtime
echo -e "\n${CYAN}[3/4] Compiling Nora Compiler and Runtime...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go (golang) is not installed or not in PATH!${NC}"
    echo -e "Please install Go to compile Nora from source, or check back once prebuilt binaries are released."
    exit 1
fi

echo -e "  - Compiling binary..."
if go build -o "$BIN_DIR/nora" pkg/cmd/nora/main.go; then
    echo -e "  ${GREEN}✓${NC} Nora compiler compiled successfully."
else
    echo -e "  ${RED}✗ Failed to compile Nora compiler!${NC}"
    exit 1
fi

echo -e "  - Copying standard library to ${YELLOW}${STD_DIR}${NC}..."
rm -rf "$STD_DIR"/*
cp -R std/* "$STD_DIR/"
echo -e "  ${GREEN}✓${NC} Standard library successfully copied."

# Step 4: Configure environment PATH
echo -e "\n${CYAN}[4/4] Integrating with environment PATH...${NC}"

PATH_LINE="export PATH=\"\$HOME/.nora/bin:\$PATH\""
SHELL_PROFILES=()

if [ -f "$HOME/.zshrc" ]; then
    SHELL_PROFILES+=("$HOME/.zshrc")
fi
if [ -f "$HOME/.bashrc" ]; then
    SHELL_PROFILES+=("$HOME/.bashrc")
fi
if [ -f "$HOME/.profile" ]; then
    SHELL_PROFILES+=("$HOME/.profile")
fi

# Fallback to .bashrc if none exist
if [ ${#SHELL_PROFILES[@]} -eq 0 ]; then
    SHELL_PROFILES+=("$HOME/.bashrc")
    touch "$HOME/.bashrc"
fi

UPDATED_ANY=false
for PROFILE in "${SHELL_PROFILES[@]}"; do
    if grep -Fq "$NORA_DIR/bin" "$PROFILE"; then
        echo -e "  - PATH is already configured in ${YELLOW}$(basename "$PROFILE")${NC}."
    else
        echo -e "  - Appending PATH to ${YELLOW}$(basename "$PROFILE")${NC}..."
        echo -e "\n# Nora Compiler & Runtime Environment" >> "$PROFILE"
        echo -e "$PATH_LINE" >> "$PROFILE"
        UPDATED_ANY=true
    fi
done

echo -e "\n${GREEN}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}         NORA LANGUAGE INSTALLED SUCCESSFULLY!${NC}"
echo -e "${GREEN}${BOLD}================================================================${NC}"
echo -e "Nora is now installed at: ${YELLOW}${NORA_DIR}${NC}"
echo ""
echo -e "To start using Nora in your current shell session, run:"
echo -e "  ${CYAN}export PATH=\"\$HOME/.nora/bin:\$PATH\"${NC}"
echo ""
if [ "$UPDATED_ANY" = true ]; then
    echo -e "For future sessions, shell configuration files have been updated."
    echo -e "Just restart your terminal or source your profile to start using the ${BOLD}nora${NC} command."
fi
echo ""
echo -e "Test your installation with:"
echo -e "  ${CYAN}nora help${NC}"
echo -e "  ${CYAN}nora targets${NC}"
echo -e "${GREEN}${BOLD}================================================================${NC}"
