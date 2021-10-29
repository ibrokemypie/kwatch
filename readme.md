# kwatch

a little tui interface to play media from a caddy fileserver.

### options:

* -a: server address
* -u: server http username
* -p: server http password

### usage:

use standard input to interact with the program. enter q to quit, p to reprint current folder listing or numbers to pick a selection from the listing.

files will be opened with mpv over http(s), folders will be entered and their contents listed for selection.
