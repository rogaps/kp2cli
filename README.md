# kp2cli

WIP

## Installing

```
go get -u github.com/rogaps/kp2cli
```

## Usage

```
$ kp2cli
kp2cli, Keepass 2 Interactive Shell
Type "help" for help.

kp2cli:/> help
cd     Change directory (path to a group)
close  Close the opened database
exit   Exit this program
find   Find entries
help   Print help
ls     List items in the pwd or specified paths
open   Open a Keepass database
xp     Copy password to clipboard
xu     Copy username to clipboard
xx     Clear the clipboard
```

## TODOs

- Implement Keyfile
- Implement `add` command
- Implement `remove` command
- Implement `move` command
- Implement generate password

## Credits

- [kpcli](http://kpcli.sourceforge.net/)
- [gokeepasslib](https://github.com/tobischo/gokeepasslib)
- [liner](https://github.com/peterh/liner)
