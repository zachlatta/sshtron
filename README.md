# ![SSHTron](https://rawgit.com/zachlatta/sshtron/master/logo.svg)

SSHTron is a multiplayer lightcycle game that runs through SSH. Just run the
command below and you'll be playing in seconds:

    $ ssh sshtron.zachlatta.com

_Controls: WASD or vim keybindings to move (**do not use your arrow keys**).
Escape or Ctrl+C to exit._

![Demo](static/img/gameplay.gif)

**Code quality disclaimer:** _SSHTron was built in ~20 hours at
[BrickHack 2](https://brickhack.io/). Here be dragons._

## Running Your Own Copy

Clone the project and `cd` into its directory. These instructions assume that
you have your `GOPATH` setup correctly.

```sh
# Create an RSA public/private keypair in the current directory for the server
# to use. Don't give it a passphrase.
$ ssh-keygen -t rsa -f id_rsa

# Download dependencies and compile the project
$ go get && go build

# Run it! You can set PORT to customize the HTTP port it serves on and SSH_PORT
# to customize the SSH port it serves on.
$ ./sshtron
```

## License

SSHTron is licensed under the MIT License. See the full license text in
[`LICENSE`](LICENSE).
