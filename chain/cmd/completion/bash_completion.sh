#!/bin/bash

# Bash completion script for aurad
# Source this file or add to your .bashrc:
#   source <(aurad completion bash)

_aurad_completion() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Root commands
    if [ $COMP_CWORD -eq 1 ]; then
        opts="init start version status keys query tx interactive config completion batch script help"
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    # Module-specific completions
    case "${prev}" in
        tx)
            local tx_modules="vcregistry vc vcr bridge br xchain inclusionroutines ir auth governance bank"
            COMPREPLY=( $(compgen -W "${tx_modules}" -- ${cur}) )
            return 0
            ;;
        query)
            local query_modules="vcregistry vc vcr bridge br xchain inclusionroutines ir auth governance bank"
            COMPREPLY=( $(compgen -W "${query_modules}" -- ${cur}) )
            return 0
            ;;
        vcregistry|vc|vcr)
            if [ "${COMP_WORDS[1]}" == "tx" ]; then
                local vc_tx="mint-vc mint create-vc revoke-vc revoke register-did reg-did create-did update-did"
                COMPREPLY=( $(compgen -W "${vc_tx}" -- ${cur}) )
            elif [ "${COMP_WORDS[1]}" == "query" ]; then
                local vc_query="vc did dids-by-controller policy policies"
                COMPREPLY=( $(compgen -W "${vc_query}" -- ${cur}) )
            fi
            return 0
            ;;
        bridge|br|xchain)
            if [ "${COMP_WORDS[1]}" == "tx" ]; then
                local bridge_tx="link-address link lock-tokens lock unlock-tokens unlock"
                COMPREPLY=( $(compgen -W "${bridge_tx}" -- ${cur}) )
            elif [ "${COMP_WORDS[1]}" == "query" ]; then
                local bridge_query="linked-addresses params"
                COMPREPLY=( $(compgen -W "${bridge_query}" -- ${cur}) )
            fi
            return 0
            ;;
        inclusionroutines|ir)
            if [ "${COMP_WORDS[1]}" == "tx" ]; then
                local ir_tx="complete submit"
                COMPREPLY=( $(compgen -W "${ir_tx}" -- ${cur}) )
            elif [ "${COMP_WORDS[1]}" == "query" ]; then
                local ir_query="ir list user-irs completion"
                COMPREPLY=( $(compgen -W "${ir_query}" -- ${cur}) )
            fi
            return 0
            ;;
        keys)
            local key_cmds="add list show delete import export mnemonic"
            COMPREPLY=( $(compgen -W "${key_cmds}" -- ${cur}) )
            return 0
            ;;
        completion)
            local shells="bash zsh fish powershell"
            COMPREPLY=( $(compgen -W "${shells}" -- ${cur}) )
            return 0
            ;;
    esac

    # Flag completions
    if [[ ${cur} == -* ]] ; then
        opts="--help --home --config --log-level --log-format --verbose --debug --output --from --gas --gas-prices --fees --chain-id --node"
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    return 0
}

complete -F _aurad_completion aurad
