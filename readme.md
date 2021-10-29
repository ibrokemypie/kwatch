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

use standard input to interact with the program. enter q to quit, p to reprint current folder listing or numbers to pick a selection from the listing.

files will be opened with the chosen application over http(s), folders will be entered and their contents listed for selection.

options can be stored in a toml config file.