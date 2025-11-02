# cradle 🧺

Install using Go,

```bash
go install github.com/gurleensethi/cradle@latest
```

Add the follwing function to your shell config file.

```bash
crdl() {
    out=$(go run . "$@")
    if [[ $out == eval* ]]; then
        eval "${out#eval}"
    else
        printf '%s\n' "$out"
    fi
}
```

Bash

```bash
curl https://github.com/gurleensethi/cradle/blob/main/crdl.sh >> ~/.bashrc
```

ZSH

```bash
curl https://github.com/gurleensethi/cradle/blob/main/crdl.sh >> ~/.zshrc
```