Output './gen/main.gif'

Source './config.tape'

# First switch: arrow down twice, then enter
Sleep 100ms
Type@100ms "crdl"
Enter
Sleep 500ms

Down 1
Sleep 250ms
Down 1
Sleep 250ms
Down 1
Sleep 750ms
Enter

Sleep 500ms

Type@150ms "pwd"
Enter

Sleep 1000ms

# Second switch: filter search, then enter the first match
Type@100ms "crdl"
Enter
Sleep 500ms

Type@100ms "/"
Type@50ms "go-playground"
Sleep 750ms
Enter
Sleep 250ms
Enter

Sleep 500ms

Type@150ms "pwd"
Enter

Sleep 1000ms

# Third switch: same arrowing as the first
Type@100ms "crdl"
Enter
Sleep 500ms

Down 1
Sleep 250ms
Down 1
Sleep 750ms
Enter

Sleep 500ms

Type@150ms "pwd"
Enter

Sleep 1000ms
