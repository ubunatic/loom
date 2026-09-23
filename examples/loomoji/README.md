# loomoji

Install the command and Zsh widget script from the repository root:

```sh
make install
```

Add the widget to Zsh by sourcing `loomoji.zsh` from `.zshrc`, then choose a
key binding:

```zsh
source ~/.local/share/loomoji/loomoji.zsh
bindkey '^X^M' loomoji-insert-widget
```

Invoke the binding while editing a command to insert the selected emoji at the
current command-line cursor. Running `loomoji` as a regular command prints the
emoji as command output instead.
