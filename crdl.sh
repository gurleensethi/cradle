crdl() {
    out=$(cradle "$@")
    if [[ $out == eval* ]]; then
        eval "${out#eval}"
    else
        printf '%s\n' "$out"
    fi
}