# Event store

A Go application to receive events as POST requests with a JSON body and store them in MongoDB.

The first use of this application is to receive [Content Security Policy reports][csp-report],
but we'd like to expand it to receive performance metrics from the frontend and any JavaScript
errors on GOV.UK.

[csp-report]: http://www.w3.org/TR/CSP2/#violation-reports

## Technical documentation

### Dependencies

- MongoDB

### Running the application

```bash
make run
```

### Running tests

```bash
make test
```

## Licence

[MIT License](LICENSE.md)
