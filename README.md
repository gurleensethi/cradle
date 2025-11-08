# Cradle 🧺

Cradle helps you manage your local projects.

## Installation

Install using **Go**:

```bash
go install github.com/gurleensethi/cradle@latest
```

## Shell Configuration

Add the helper function to your shell's configuration file:

```bash
crdl() {
    out=$(CRADLE_CMDOUT=1 cradle "$@" 3>&1 1>&2 2>&3)
    if [[ $out == eval* ]]; then
        eval "${out#eval}"
    else
        printf '%s' "$out"
    fi
}
```

> **Note**: This configuration is required to enable directory switching functionality. Without it, you won't be able to change directories using Cradle.

### For Bash Users

To configure Cradle for Bash, append the helper script to your `.bashrc`:

```bash
curl -s https://raw.githubusercontent.com/gurleensethi/cradle/main/crdl.sh >> ~/.bashrc
```

### For Zsh Users

To configure Cradle for Zsh, append the helper script to your `.zshrc`:

```bash
curl -s https://raw.githubusercontent.com/gurleensethi/cradle/main/crdl.sh >> ~/.zshrc
```

After adding the script, remember to restart your terminal or source the configuration file to apply the changes.

## Usage

### Use with TUI.

![](./docs/vhs/gen/main.gif)

### List projects

![](./docs/vhs/gen/list.gif)