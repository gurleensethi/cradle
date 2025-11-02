# cradle 🧺

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
