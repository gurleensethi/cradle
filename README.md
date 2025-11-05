# Cradle 🧺

Cradle helps you manage your local projects.

## Installation

To install Cradle, use the following command:

```bash
go install github.com/gurleensethi/cradle@latest
```

## Shell Configuration

Add the helper function to your shell's configuration file for easy usage:

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
