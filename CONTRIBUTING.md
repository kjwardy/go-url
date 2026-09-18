# Contributing

Contributions are welcome, whether they are bug reports, feature ideas, documentation improvements, or code changes.

## Raising an issue

Please raise an issue if you find a bug, have an idea for a feature, or think part of the project could be improved. Issues are useful even when you do not have time to implement the change yourself.

Before opening a new issue, check the existing issues to avoid duplicates. When reporting a bug, include:

- A clear description of the problem and the expected behavior
- Steps that reliably reproduce the problem
- Relevant logs or error messages, with credentials and other sensitive information removed
- Your operating system, browser, and application version where applicable

For feature requests, explain the problem the feature would solve and provide an example of how it might work if possible.

## Submitting a change

1. Fork this repo into your own GitHub space.
1. Create a branch from the latest `master` branch using a descriptive name such as `fix/delete-dialog` or `feature/more-query-metrics`.
1. Make focused changes that follow the conventions used by the existing code.
1. Add or update tests where appropriate.
1. Run the relevant checks described below.
1. Commit your changes with a concise message explaining their purpose.
1. Push the branch to your fork and open a pull request against this repository's `master` branch.

Keep pull requests focused on one change where possible. In the pull request description, explain what changed, why it was needed, how it was tested, and link any related issues. Screenshots are encouraged for visible frontend changes.

Draft pull requests are welcome when you would like early feedback on an approach.

## Development setup

The application requires Node.js, Yarn, Go, and PostgreSQL. On macOS, the required development tools can be installed with:

```sh
brew install nvm yarn go
nvm install
```

Start PostgreSQL manually or with Docker Compose:

```sh
docker-compose up postgres
```

### Start the frontend

```sh
cd frontend
yarn
yarn start
```

### Start the API

In a separate terminal:

```sh
cd api
go mod download
POSTGRES_PASS=password HOSTS=localhost APP_URI=http://localhost:3000 go run server.go
```

Open http://localhost:3000/go after the services have started.

## Verification

Run the Go test suite from the API directory:

```sh
cd api
go test ./...
```

Run the frontend formatting, lint, and production build checks from the frontend directory:

```sh
cd frontend
yarn prettier:check
yarn lint
yarn build
```

Please resolve new warnings or failures caused by your changes before submitting a pull request.

## Pull request review

Maintainers may request changes or ask questions during review. Please keep discussion constructive and update the pull request when feedback is addressed. Once approved, a maintainer will merge the change.
