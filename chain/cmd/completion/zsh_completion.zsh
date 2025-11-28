#compdef aurad

# Zsh completion script for aurad
# Install: aurad completion zsh > "${fpath[1]}/_aurad"

_aurad() {
    local -a commands

    _arguments -C \
        '--help[Show help]' \
        '--home[Directory for config and data]:directory:_files -/' \
        '--config[Config file]:file:_files' \
        '--log-level[Logging level]:level:(trace debug info warn error fatal panic)' \
        '--log-format[Logging format]:format:(json plain)' \
        '--verbose[Enable verbose output]' \
        '--debug[Enable debug output]' \
        '--output[Output format]:format:(text json yaml csv)' \
        '1: :->cmds' \
        '*:: :->args'

    case $state in
        cmds)
            commands=(
                'init:Initialize node configuration'
                'start:Start the blockchain node'
                'version:Display version information'
                'status:Show node status'
                'keys:Manage keys'
                'query:Query blockchain state'
                'tx:Create and sign transactions'
                'interactive:Start interactive mode'
                'config:Manage configuration'
                'completion:Generate shell completion'
                'batch:Execute batch commands'
                'script:Execute script file'
            )
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                tx)
                    _aurad_tx
                    ;;
                query)
                    _aurad_query
                    ;;
                keys)
                    _aurad_keys
                    ;;
                completion)
                    _values 'shell' bash zsh fish powershell
                    ;;
            esac
            ;;
    esac
}

_aurad_tx() {
    local -a modules
    modules=(
        'vcregistry:VC Registry transactions'
        'vc:VC Registry transactions (alias)'
        'vcr:VC Registry transactions (alias)'
        'bridge:Bridge transactions'
        'br:Bridge transactions (alias)'
        'xchain:Bridge transactions (alias)'
        'inclusionroutines:Inclusion Routines transactions'
        'ir:Inclusion Routines transactions (alias)'
        'auth:Auth module transactions'
        'governance:Governance transactions'
        'bank:Bank module transactions'
    )

    if (( CURRENT == 2 )); then
        _describe 'module' modules
    elif (( CURRENT == 3 )); then
        case $words[2] in
            vcregistry|vc|vcr)
                _values 'vc command' \
                    'mint-vc[Mint a new VC]' \
                    'mint[Mint a new VC (alias)]' \
                    'create-vc[Mint a new VC (alias)]' \
                    'revoke-vc[Revoke a VC]' \
                    'revoke[Revoke a VC (alias)]' \
                    'register-did[Register a DID]' \
                    'reg-did[Register a DID (alias)]' \
                    'create-did[Register a DID (alias)]' \
                    'update-did[Update DID document]'
                ;;
            bridge|br|xchain)
                _values 'bridge command' \
                    'link-address[Link addresses]' \
                    'link[Link addresses (alias)]' \
                    'link-addr[Link addresses (alias)]' \
                    'lock-tokens[Lock tokens]' \
                    'lock[Lock tokens (alias)]' \
                    'unlock-tokens[Unlock tokens]' \
                    'unlock[Unlock tokens (alias)]'
                ;;
            inclusionroutines|ir)
                _values 'ir command' \
                    'complete[Complete an IR]' \
                    'submit[Submit IR completion]'
                ;;
        esac
    fi
}

_aurad_query() {
    local -a modules
    modules=(
        'vcregistry:Query VC Registry'
        'vc:Query VC Registry (alias)'
        'vcr:Query VC Registry (alias)'
        'bridge:Query Bridge'
        'br:Query Bridge (alias)'
        'xchain:Query Bridge (alias)'
        'inclusionroutines:Query Inclusion Routines'
        'ir:Query Inclusion Routines (alias)'
        'auth:Query Auth module'
        'governance:Query Governance'
        'bank:Query Bank module'
    )

    if (( CURRENT == 2 )); then
        _describe 'module' modules
    elif (( CURRENT == 3 )); then
        case $words[2] in
            vcregistry|vc|vcr)
                _values 'query' \
                    'vc[Query VC by ID]' \
                    'did[Query DID by ID]' \
                    'dids-by-controller[Query DIDs by controller]' \
                    'policy[Query VC policy]' \
                    'policies[List all policies]'
                ;;
            bridge|br|xchain)
                _values 'query' \
                    'linked-addresses[Query linked addresses]' \
                    'params[Query bridge parameters]'
                ;;
            inclusionroutines|ir)
                _values 'query' \
                    'ir[Query IR by ID]' \
                    'list[List all IRs]' \
                    'user-irs[Query user IRs]' \
                    'completion[Query IR completion]'
                ;;
        esac
    fi
}

_aurad_keys() {
    _values 'key command' \
        'add[Add a new key]' \
        'list[List all keys]' \
        'show[Show key details]' \
        'delete[Delete a key]' \
        'import[Import a key]' \
        'export[Export a key]' \
        'mnemonic[Generate mnemonic]'
}

_aurad "$@"
