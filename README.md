# Vinyl Scanner

The vinyl scanner is a project we started with the idea of tracking our vinyl collection, as well as tracking the times we play our albums. Think of Last.fm, but handmade and self-hosted. We have it hanging above the record player.

It is composed by two parts: a wooden shelf with an ESP32 device, and a server program that runs somewhere else. The shelf holds the vinyl while we're playing it, and the ESP32 inside contacts the server to track the vinyl we're playing. Therefore, each vinyl must have its own NFC tag in order to be trackable.

## Clone

This repository contains Git submodules. Therefore, when cloning it, don't
forget to clone it with submodules:

```shell
git clone --recurse-submodules https://github.com/ChrisB2351/vinyl-scanner
```

## Shelf

The shelf was made out of 3mm multiplex wood and the "Playing Now" sign out of acrylic. The schematics for laser cutting can be found in the following files:

- [`docs/shelf.pdf`](./docs/shelf.pdf): ready to print PDF file exported to life-size.
- [`docs/shelf.dwg`](./docs/shelf.dwg): original file with the schematics and a layer with the dimensions. 

## Documentation

You can find more documentation in the following files:

- [`arduino`](./arduino/README.md) includes more information regarding the ESP32 module.
- [`server`](./server/README.md) includes more information regarding the server software.

## License

[MIT License](./LICENSE)
