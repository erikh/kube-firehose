# kube-firehose: replay all container logs from your cluster on stdout

A very simple program to yield the stdout logs from all containers running in the cluster. Logs are annotated with where they came from, and pod volatility is accounted for: new pods will pick up as they are available, and old pods will expire.

**NOTE** by default **all log history is replayed**. _This is probably not what you want on a long-running cluster!_ Pass the `-t` option to just show current events. You've been warned! :)

## Installation

```
go install github.com/erikh/kube-firehose@latest
```

## Usage

Options:

Durations are Golang durations, such as `1m` or `1h15s`.

-   `-t`: Just current events, no history
-   `-since <duration>`: Show log messages that were written in the last `<duration>`.
-   `-wait <duration>`: Wait this amount of time, playing logs, and then terminate the program.
- `-kubeconfig`: Provide your Kubernetes configuration. Uses `~/.kube/config` by default.

## License

MIT

## Author

Erik Hollensbe <git@hollensbe.org>
