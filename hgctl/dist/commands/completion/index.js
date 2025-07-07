"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupCompletionCommand = setupCompletionCommand;
function setupCompletionCommand(parent) {
    parent
        .command('completion')
        .description('Generate shell completion script')
        .argument('[shell]', 'Shell type (bash|zsh|fish)', 'bash')
        .action((shell) => {
        switch (shell) {
            case 'bash':
                console.log(getBashCompletion());
                break;
            case 'zsh':
                console.log(getZshCompletion());
                break;
            case 'fish':
                console.log(getFishCompletion());
                break;
            default:
                console.error(`Unsupported shell: ${shell}`);
                console.error('Supported shells: bash, zsh, fish');
                process.exit(1);
        }
    });
}
function getBashCompletion() {
    return `#!/usr/bin/env bash
# hgctl bash completion script
# Add this to your ~/.bashrc or ~/.bash_profile:
# source <(hgctl completion bash)

_hgctl_completion() {
  local cur prev words cword
  _init_completion || return

  local commands="get deploy remove context completion"
  local get_commands="performers releases"
  local deploy_commands="artifact"
  local remove_commands="performer"
  local context_commands="set show use list"

  case "\${words[1]}" in
    get)
      if [[ \${cword} -eq 2 ]]; then
        COMPREPLY=( \$(compgen -W "\${get_commands}" -- \${cur}) )
      fi
      ;;
    deploy)
      if [[ \${cword} -eq 2 ]]; then
        COMPREPLY=( \$(compgen -W "\${deploy_commands}" -- \${cur}) )
      fi
      ;;
    remove|rm|delete)
      if [[ \${cword} -eq 2 ]]; then
        COMPREPLY=( \$(compgen -W "\${remove_commands}" -- \${cur}) )
      fi
      ;;
    context)
      if [[ \${cword} -eq 2 ]]; then
        COMPREPLY=( \$(compgen -W "\${context_commands}" -- \${cur}) )
      fi
      ;;
    *)
      if [[ \${cword} -eq 1 ]]; then
        COMPREPLY=( \$(compgen -W "\${commands}" -- \${cur}) )
      fi
      ;;
  esac
}

complete -F _hgctl_completion hgctl`;
}
function getZshCompletion() {
    return `#compdef hgctl
# hgctl zsh completion script
# Add this to your ~/.zshrc:
# source <(hgctl completion zsh)

_hgctl() {
  local -a commands
  commands=(
    'get:Get resources'
    'deploy:Deploy resources'
    'remove:Remove resources'
    'context:Manage contexts'
    'completion:Generate shell completion script'
  )

  local -a get_commands
  get_commands=(
    'performers:List all performers'
    'releases:List releases for an AVS'
  )

  local -a deploy_commands
  deploy_commands=(
    'artifact:Deploy an artifact for an AVS'
  )

  local -a remove_commands
  remove_commands=(
    'performer:Remove a performer'
  )

  local -a context_commands
  context_commands=(
    'set:Set context configuration values'
    'show:Show context configuration'
    'use:Switch to a different context'
    'list:List all contexts'
  )

  _arguments -C \\
    '1: :->command' \\
    '2: :->subcommand' \\
    '*::arg:->args'

  case $state in
    command)
      _describe 'command' commands
      ;;
    subcommand)
      case \$words[1] in
        get)
          _describe 'get command' get_commands
          ;;
        deploy)
          _describe 'deploy command' deploy_commands
          ;;
        remove|rm|delete)
          _describe 'remove command' remove_commands
          ;;
        context)
          _describe 'context command' context_commands
          ;;
      esac
      ;;
  esac
}

_hgctl "$@"`;
}
function getFishCompletion() {
    return `# hgctl fish completion script
# Add this to your ~/.config/fish/completions/hgctl.fish

# Disable file completions
complete -c hgctl -f

# Main commands
complete -c hgctl -n "__fish_use_subcommand" -a "get" -d "Get resources"
complete -c hgctl -n "__fish_use_subcommand" -a "deploy" -d "Deploy resources"
complete -c hgctl -n "__fish_use_subcommand" -a "remove" -d "Remove resources"
complete -c hgctl -n "__fish_use_subcommand" -a "context" -d "Manage contexts"
complete -c hgctl -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion script"

# Get subcommands
complete -c hgctl -n "__fish_seen_subcommand_from get; and not __fish_seen_subcommand_from performers releases" -a "performers" -d "List all performers"
complete -c hgctl -n "__fish_seen_subcommand_from get; and not __fish_seen_subcommand_from performers releases" -a "releases" -d "List releases for an AVS"

# Deploy subcommands
complete -c hgctl -n "__fish_seen_subcommand_from deploy; and not __fish_seen_subcommand_from artifact" -a "artifact" -d "Deploy an artifact for an AVS"

# Remove subcommands
complete -c hgctl -n "__fish_seen_subcommand_from remove; and not __fish_seen_subcommand_from performer" -a "performer" -d "Remove a performer"
complete -c hgctl -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from performer" -a "performer" -d "Remove a performer"
complete -c hgctl -n "__fish_seen_subcommand_from delete; and not __fish_seen_subcommand_from performer" -a "performer" -d "Remove a performer"

# Context subcommands
complete -c hgctl -n "__fish_seen_subcommand_from context; and not __fish_seen_subcommand_from set show use list" -a "set" -d "Set context configuration values"
complete -c hgctl -n "__fish_seen_subcommand_from context; and not __fish_seen_subcommand_from set show use list" -a "show" -d "Show context configuration"
complete -c hgctl -n "__fish_seen_subcommand_from context; and not __fish_seen_subcommand_from set show use list" -a "use" -d "Switch to a different context"
complete -c hgctl -n "__fish_seen_subcommand_from context; and not __fish_seen_subcommand_from set show use list" -a "list" -d "List all contexts"`;
}
//# sourceMappingURL=index.js.map