# kwatch

a little tui interface to play media from a caddy fileserver.

### options:

* -a: server address
* -u: server http username
* -p: server http password
* -o: program to open files with (defaults to mpv)
* -w: write current options to config file
* -c: set config file to use (defaults to os.UserConfigDir()/kwatch.toml)

### usage:

pick directories to explore or files to open with the keyboard, selecting with enter. `/` can be used to fuzzy find in the list.

files will be opened with the chosen application over http(s), folders will be entered and their contents listed for selection.

options can be stored in a toml config file.