# go-start-stop

go-start-stop demonstrates a context based service start-stop pattern. It's not
a library, just a starting point for your own services.

"A little copying is better than a little dependency."
--Rob Pike, [Go Proverbs](https://go-proverbs.github.io/)

## Usage

- `go run main.go` will demonstrate services failing and being restarted.
- `go run main.go -clean` will prevent services from failing to demonstrate
  signal handling, i.e. `Ctrl-C`.

## License

Public domain.
