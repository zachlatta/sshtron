# ![SSHTron](https://cdn.rawgit.com/zachlatta/sshtron/master/logo.svg)

SSHTron is a multiplayer lightcycle game that runs through SSH. Just run the command below and you'll be playing in seconds:

    $ ssh sshtron.zachlatta.com

_Controls: WASD or vim keybindings to move (**do not use your arrow keys**). Escape or Ctrl+C to exit._

![Demo](static/img/gameplay.gif)

**Code quality disclaimer:** _SSHTron was built in ~20 hours at [BrickHack 2](https://brickhack.io/). Here be dragons._

## Want to choose color yourself?

There are total 7 colors to choose from: Red, Green, Yellow, Blue, Magenta, Cyan and White

    $ ssh red@sshtron.zachlatta.com

If the color you picked is already taken in all open games, you'll randomly be assigned a color.

## Running Your Own Copy

Clone the project and `cd` into its directory. These instructions assume that you have your `GOPATH` setup correctly.

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

## Running under a Docker container

Clone the project and `cd` into its directory.

```sh
# Build the SSHTron Docker image
$ docker build -t sshtron .

# Spin up the container with always-restart policy
$ docker run -t -d -p 2022:2022 --restart always --name sshtron sshtron
```

For Raspberry Pi, use the following to build the Docker image:

```sh
$ docker build -t sshtron --build-arg BASE_IMAGE=resin/raspberry-pi-golang:latest .
```

## CVE-2016-0777

[CVE-2016-0777](https://www.qualys.com/2016/01/14/cve-2016-0777-cve-2016-0778/openssh-cve-2016-0777-cve-2016-0778.txt)
revealed two SSH client vulnerabilities that can be exploited by a malicious SSH server. While SSHTron does not exploit
these vulnerabilities, you should still patch your client before you play. SSHTron is open source, but the server
could always be running a modified version of SSHTron that does exploit the vulnerabilities described
in [CVE-2016-0777](https://www.qualys.com/2016/01/14/cve-2016-0777-cve-2016-0778/openssh-cve-2016-0777-cve-2016-0778.txt).

If you haven't yet patched your SSH client, you can follow
[these instructions](https://www.jacobtomlinson.co.uk/quick%20tip/2016/01/15/fixing-ssh-vulnerability-CVE-2016-0777/) to do so now.

## License

SSHTron is licensed under the MIT License. See the full license text in [`LICENSE`](LICENSE).
