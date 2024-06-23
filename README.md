# Charm Chatroom SSH Thing
A small project using some [charmbracelet](https://github.com/charmbracelet) packages ([bubbletea](https://github.com/charmbracelet/bubbletea) and [wish](https://github.com/charmbracelet/wish) so far). This was created to try out the charm stuff and as a personal project for the [boot.dev](www.boot.dev) course

## Usage
Build and run the main.go file passing in the ./test.db (or whichever sqlite db file) and then connect over ssh  
`ssh {HOST} -p {PORT} -l {HOSTNAME}`

![Screenshot from 2024-06-21 23-29-23](https://github.com/jpleatherland/chatroom/assets/19578072/41de0c4c-9884-45d0-8194-9861b03c44dd)

## TODO
- Make it pretty with Lipgloss etc
- See how it works with multiple concurrent users.
