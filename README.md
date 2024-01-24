A very simple program to yield the stdout logs from all containers running in the cluster. Logs are annotated with where they came from, and pod volatility is accounted for: new pods will pick up as they are available, and old pods will expire.

**NOTE** by default **all log history is replayed**. _This is probably not what you want on a long-running cluster!_ Pass the `-t` option to just show current events. You've been warned! :)

You can also pass `-since` with a Golang duration to review items that occurred in that amount of time.

Supply `-wait <duration>` to wait for that amount of time before terminating the program. Combine with `-since` to scope your logs in digestible sizes.

Usage is simple: `kube-firehose`. If you have a configuration you want to use, specify `-kubeconfig`. It uses `~/.kube/config` by default.

License is MIT, Erik Hollensbe <git@hollensbe.org> is the author.
