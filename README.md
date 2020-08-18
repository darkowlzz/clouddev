# clouddev

Go implementation of [remotedev](https://github.com/darkowlzz/remotedev).

## Development

- Build the binary with `make clouddev`.
- Run all the tests with `make test`.
- Update go dependencies with `make tidy`.
- Lint code with `make lint`.
- Add new command with `make cobra ARGS="add cmdX"`.
- Run the above in docker by adding `-docker` suffix, e.g. `make tidy-docker`.
