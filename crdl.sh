
crdl() {
    out=$(CRADLE_CMDOUT=1 cradle "$@" 3>&1 1>&2 2>&3)
    if [[ $out == eval* ]]; then
        eval "${out#eval}"
    else
        printf '%s\n' "$out"
    fi
}