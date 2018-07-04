# ASharedJourney

This is a fork of a project originally developed during the SFR game jam along side Pierre, Gabriel, Aurore & Fabio.

This project uses the go [pixel](https://github.com/faiface/pixel) package for sound management and sprites, go check them out.


>GameJam SFR 2018 Julia - Pierre - Gabriel - Aurore - Fabio
>
>Music: Thibault
>
>Theme: Si j'étais toi et que tu étais moi (If I were you and you were me)
>
>[Itch.io](https://fmaschi.itch.io/a-shared-journey)

## Getting Started

* [GO](https://golang.org) - Programming language

## Building and running

### Installation

- First, install the game and its dependencies

```bash
go get -u github.com/gandrin/ASharedJourney
```

- You will also need the `go-bindata` program to build the assets into the binary file

```bash
go get -u github.com/jteeuwen/go-bindata/...
```

> Make sure your `$GOPATH` is set :wink:

### Run

```
make run
```

### Releasing

```
make build_mac
```
#### OR 
```
make build_linux
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

### Acknowledgements

Thibault A. - Sound designer 