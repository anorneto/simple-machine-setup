# Source Aliases
source ~/.config/fish/alias.fish

if status is-interactive
    # Commands to run in interactive sessions can go here
end

starship init fish | source
fzf --fish | source
zoxide init fish | source
~/.local/bin/mise activate fish | source
