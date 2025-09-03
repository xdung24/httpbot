uinput-go


Notes:
- Requires Linux or WSL to compile

## Build

To build the project, install golang:
```sh
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
# Add to PATH to .profile or .bashrc
export PATH=$PATH:/usr/local/go/bin
```

To build this app, linux environment is required because linux head import

```sh
sudo apt-get update && sudo apt-get install -y build-essential linux-headers-generic gcc-multilib libc6-dev-i386
```