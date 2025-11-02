crdl() {
    local out="$(go run . $@)"
    if [[ "$out" == "eval"* ]]; then
        eval $(echo $out | cut -d' ' -f2-)
    else
        echo $out
    fi
}
