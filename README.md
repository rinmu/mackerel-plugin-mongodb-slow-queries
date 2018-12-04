# mackerel-plugin-mongodb-slow-queries

Sample plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-mongodb-slow-queries [-metric-key-prefix=<prefix>][-host=<host>] [-port=<port>] [-username=<username>] [-password=<password>] [-database=<database>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.sample]
command = "/path/to/mackerel-plugin-mongodb-slow-queries"
```
