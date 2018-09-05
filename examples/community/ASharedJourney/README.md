# ASharedJourney

This is a fork of a project originally developed during the SFR game jam along side Pierre, Gabriel, Aurore & Fabio.
Original repo can be found [here](https://github.com/gandrin/ASharedJourney).
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

- You will need the `go-bindata` program to build the assets into the binary file

```bash
go get -u github.com/jteeuwen/go-bindata/...
```

> Make sure your `$GOPATH` is set to the root of pixel :wink:

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