# ptsd shell completion for fish
# Install: ptsd completion fish | source

function __ptsd_features
    ptsd feature list --agent 2>/dev/null | awk '{print $1}'
end

function __ptsd_no_subcommand
    set -l cmd (commandline -opc)
    test (count $cmd) -eq 1
end

function __ptsd_using_command
    set -l cmd (commandline -opc)
    test (count $cmd) -ge 2; and test $cmd[2] = $argv[1]
end

function __ptsd_using_subcommand
    set -l cmd (commandline -opc)
    test (count $cmd) -ge 3; and test $cmd[2] = $argv[1]; and test $cmd[3] = $argv[2]
end

function __ptsd_needs_subcommand
    set -l cmd (commandline -opc)
    test (count $cmd) -eq 2; and test $cmd[2] = $argv[1]
end

# Top-level commands
complete -c ptsd -n __ptsd_no_subcommand -f -a "init adopt feature config task prd seed bdd test status validate hooks review skills issues context gate-check auto-track help version completion"

# feature
complete -c ptsd -n "__ptsd_needs_subcommand feature" -f -a "add list status show remove"
complete -c ptsd -n "__ptsd_using_subcommand feature show" -f -a "(__ptsd_features)"
complete -c ptsd -n "__ptsd_using_subcommand feature status" -f -a "(__ptsd_features)"
complete -c ptsd -n "__ptsd_using_subcommand feature remove" -f -a "(__ptsd_features)"

# task
complete -c ptsd -n "__ptsd_needs_subcommand task" -f -a "add list next done update"
complete -c ptsd -n "__ptsd_using_subcommand task add" -f -a "(__ptsd_features)"

# prd
complete -c ptsd -n "__ptsd_needs_subcommand prd" -f -a "check show"
complete -c ptsd -n "__ptsd_using_subcommand prd show" -f -a "(__ptsd_features)"

# seed
complete -c ptsd -n "__ptsd_needs_subcommand seed" -f -a "init add"
complete -c ptsd -n "__ptsd_using_subcommand seed init" -f -a "(__ptsd_features)"
complete -c ptsd -n "__ptsd_using_subcommand seed add" -f -a "(__ptsd_features)"

# bdd
complete -c ptsd -n "__ptsd_needs_subcommand bdd" -f -a "add list"
complete -c ptsd -n "__ptsd_using_subcommand bdd add" -f -a "(__ptsd_features)"
complete -c ptsd -n "__ptsd_using_subcommand bdd list" -f -a "(__ptsd_features)"

# test
complete -c ptsd -n "__ptsd_needs_subcommand test" -f -a "run map"
complete -c ptsd -n "__ptsd_using_subcommand test run" -f -a "(__ptsd_features)"

# hooks
complete -c ptsd -n "__ptsd_needs_subcommand hooks" -f -a "install validate-commit pre-tool-use post-tool-use"

# review — feature ID then stage
complete -c ptsd -n "__ptsd_needs_subcommand review" -f -a "(__ptsd_features) gate"
complete -c ptsd -n "__ptsd_using_command review" -f -a "prd seed bdd tests impl"

# skills
complete -c ptsd -n "__ptsd_needs_subcommand skills" -f -a "generate generate-all list"

# issues
complete -c ptsd -n "__ptsd_needs_subcommand issues" -f -a "add list remove"

# config
complete -c ptsd -n "__ptsd_needs_subcommand config" -f -a "show"

# completion
complete -c ptsd -n "__ptsd_needs_subcommand completion" -f -a "bash fish"

# Global flag
complete -c ptsd -l agent -d "Machine-readable output"
