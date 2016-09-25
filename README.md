[![Build Status](https://travis-ci.org/gummiboll/hing.svg?branch=master)](https://travis-ci.org/gummiboll/hing)

hing
=======

`hing` is a small http pinger designed to run on your local computer to help with debugging of slow web sites.

# Install/setup
`go get github.com/gummiboll/hing` and copy bin/hing to your $PATH. Or download the [latest binary release](https://github.com/gummiboll/hing/releases/latest)

# Usage

Just run `hing` and point your browser to http://localhost:8080 (listen address can be changed with -l and port with -p).

# Developing
- runt hing with `-dev` in the same dir as the static-folder to be able to edit html/css/js without rebuilding with statik.
- build statik with `$GOPATH/bin/statik -src=static`
- Run `bra run` to enable rebuild/reload automatically after save

# Todo
- Better ui
- Tests
- Cleanup

Pull requests are more then welcome, especially for the UI.
