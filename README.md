# k8sinfo

Continously scrape one or more kubernetes clusters and make gathered information available via a simple http json get interface-
Infos scraped: `Context`, `Name`, `Namespace`, `Image` for each Container or InitContainer for any Deployment found in all namespaces.

```console
usage: k8sinfo [<flags>]

Flags:
      --help              Show context-sensitive help (also try --help-long and --help-man).
  -c, --kubeconfig=~/.kube/config ...
                          path to one or multiple kubeconfig files
  -i, --scrapeInterval=2  Interval between data scraping
  -l, --web.listen-address=":2112"
                          Address to listen on for http requests
  -d, --debug             Set log level to debug
  ```
