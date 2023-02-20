# go-pkg-server


## What is it?

Package server for golang.unexpl0.red.
Inspired by: https://git.zx2c4.com/golang-package-server/about/


Redirects here for now, because I'm scared of the Great Google Cannon(https://news.ycombinator.com/item?id=34310674)


## Why?
Because I am vain, and I want vanity imports :^).

## Features

- automatic TLS certificate management/provisioning
  - certificates fetched from LetsEncrypt with `golang.org/x/crypto/autocert`

- compiles to static executable `CGO_ENABLED=0`
  - bundles the accompanying `pkgs.txt` file inside the binary for ease of deployment

- drops privleges
  - takes flags `uid` and `gid` to enable dropping of priveleges after the priveleged bind on port 443 has occured

## How to use

change the associations in pkgs.txt, and also change the `hostname` constant in `main.go`, run go build and get those sweet vanity imports!

