# Charm Chatroom SSH Thing
A small project using some [charmbracelet](https://github.com/charmbracelet) packages ([bubbletea](https://github.com/charmbracelet/bubbletea) and [wish](https://github.com/charmbracelet/wish) so far). This was created to try out the charm stuff and as a personal project for the [boot.dev](https://www.boot.dev) course

## Usage
Create a config.json in the root directory
```
{
  "host": "localhost",
  "port": "22",
  "databaseFile": "/path/to/db.db",
  "hostKeyPath": "/path/to/sshKey"
}
```

Build and run the main.go file. 
Connect to the server over ssh e.g.  
`ssh {HOST} -p {PORT} -l {HOSTNAME} -i /path/to/ed25519 public key`

![Screenshot from 2024-06-21 23-29-23](https://github.com/jpleatherland/chatroom/assets/19578072/41de0c4c-9884-45d0-8194-9861b03c44dd)

## Supporting database
For the chat history table:  
`CREATE TABLE ChatHistory(id INTEGER, User TEXT, TimeStamp TEXT, message TEXT, PRIMARY KEY(id))`

For the users table:  
`CREATE TABLE Users (userid INTEGER PRIMARY KEY, username TEXT, publickey BLOB UNIQUE, authorised INTEGER)`

## TODO
- Make it pretty with Lipgloss etc
- Figure out how to expose this to the internet safely
- See how it works with multiple concurrent users.
