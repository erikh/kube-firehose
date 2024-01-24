A very simple program to yield the stdout logs from all containers running in the cluster. Logs are annotated with where they came from, and pod volatility is accounted for: new pods will pick up as they are available, and old pods will expire.

Usage is simple: `kube-firehose`. If you have a configuration you want to use, specify `-kubeconfig`. It uses `~/.kube/config` by default.

License is MIT, Erik Hollensbe <git@hollensbe.org> is the author.
