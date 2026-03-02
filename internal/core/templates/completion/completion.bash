#!/bin/bash
# ptsd shell completion for bash
# Install: eval "$(ptsd completion bash)"

_ptsd_features() {
    ptsd feature list --agent 2>/dev/null | awk '{print $1}'
}

_ptsd_completions() {
    local cur prev cmd
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Position 1: top-level commands
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        local commands="init adopt feature config task prd seed bdd test status validate hooks review skills issues context gate-check auto-track help version completion"
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
        return
    fi

    cmd="${COMP_WORDS[1]}"

    # Position 2: subcommands
    if [[ ${COMP_CWORD} -eq 2 ]]; then
        case "${cmd}" in
            feature)
                COMPREPLY=($(compgen -W "add list status show remove" -- "${cur}"))
                return ;;
            task)
                COMPREPLY=($(compgen -W "add list next done update" -- "${cur}"))
                return ;;
            prd)
                COMPREPLY=($(compgen -W "check show" -- "${cur}"))
                return ;;
            seed)
                COMPREPLY=($(compgen -W "init add" -- "${cur}"))
                return ;;
            bdd)
                COMPREPLY=($(compgen -W "add list" -- "${cur}"))
                return ;;
            test)
                COMPREPLY=($(compgen -W "run map" -- "${cur}"))
                return ;;
            hooks)
                COMPREPLY=($(compgen -W "install validate-commit pre-tool-use post-tool-use" -- "${cur}"))
                return ;;
            skills)
                COMPREPLY=($(compgen -W "generate generate-all list" -- "${cur}"))
                return ;;
            issues)
                COMPREPLY=($(compgen -W "add list remove" -- "${cur}"))
                return ;;
            config)
                COMPREPLY=($(compgen -W "show" -- "${cur}"))
                return ;;
            completion)
                COMPREPLY=($(compgen -W "bash fish" -- "${cur}"))
                return ;;
        esac
    fi

    # Position 3+: feature IDs where appropriate
    local sub="${COMP_WORDS[2]}"
    case "${cmd}" in
        feature)
            if [[ "${sub}" =~ ^(show|status|remove)$ && ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "$(_ptsd_features)" -- "${cur}"))
                return
            fi ;;
        review|prd)
            if [[ ${COMP_CWORD} -eq 2 ]] || [[ "${sub}" == "show" && ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "$(_ptsd_features)" -- "${cur}"))
                return
            fi ;;
        seed|bdd)
            if [[ "${sub}" =~ ^(init|add|list)$ && ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "$(_ptsd_features)" -- "${cur}"))
                return
            fi ;;
        test)
            if [[ "${sub}" == "run" && ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "$(_ptsd_features)" -- "${cur}"))
                return
            fi
            if [[ "${sub}" == "map" ]]; then
                COMPREPLY=($(compgen -f -- "${cur}"))
                return
            fi ;;
        task)
            if [[ "${sub}" == "add" && ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "$(_ptsd_features)" -- "${cur}"))
                return
            fi ;;
    esac

    # review: stage completion at position 3
    if [[ "${cmd}" == "review" && ${COMP_CWORD} -eq 3 ]]; then
        COMPREPLY=($(compgen -W "prd seed bdd tests impl" -- "${cur}"))
        return
    fi

    # Fallback: --agent flag
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--agent" -- "${cur}"))
    fi
}

complete -F _ptsd_completions ptsd
