# cosmos-node-exporter

![Latest release](https://img.shields.io/github/v/release/QuokkaStake/cosmos-node-exporter)
[![Actions Status](https://github.com/QuokkaStake/cosmos-node-exporter/workflows/test/badge.svg)](https://github.com/QuokkaStake/cosmos-node-exporter/actions)
[![codecov](https://codecov.io/gh/QuokkaStake/cosmos-node-exporter/graph/badge.svg?token=O7WAAKM6YM)](https://codecov.io/gh/QuokkaStake/cosmos-node-exporter)

cosmos-node-exporter is a Prometheus scraper that scrapes some data to monitor your node.
It exposes the following metrics:
- node status (voting power, whether the node is catching up or is stuck behind the blockchain)
- app version (local binary, latest GitHub/Gitopia release and if you are running the latest version)
- Cosmovisor metrics (version of Cosmovisor version itself)
- upgrades metrics (time till upgrade, upgrade version, if you have a binary prepared for the upgrade)
- chain metrics (cosmos-sdk version, Tendermint/CometBFT version, Go version/build tags)
- node params (minimum-gas-prices)

Specifically, if you are a validator or a node operator, you can set up alerting if:
- your app version does not match the latest on GitHub (can be useful to be notified on new releases)
- your voting power is 0 for a validator node
- your node is catching up
- there are chain upgrades your node does not have binaries for
- there's an upgrade coming soon

## How can I set it up?

First, you need to download the latest release from [the releases page](https://github.com/QuokkaStake/cosmos-node-exporter/releases/).
After that, you should unzip it, and you are ready to go:

```sh
wget <the link from the releases page>
tar <the filename you've just downloaded>
./cosmos-node-exporter <params>
```

Alternatively, install `golang` (>1.18), clone the repo and build it:
```
git clone https://github.com/QuokkaStake/cosmos-node-exporter
cd cosmos-node-exporter
# This will generate a `cosmos-node-exporter` binary file in the repository folder
make build
# This will generate a `missed-blocks-checker` binary file in $GOPATH/bin
```

To run it in detached mode in background, first, we have to copy the file to the system apps folder:

```sh
sudo cp ./cosmos-node-exporter /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/cosmos-node-exporter.service
```

You can use this template (change the user to whatever user you want this to be executed from.
It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Cosmos Node Exporter
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=cosmos-node-exporter --config <path to config>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to autostart and run it:

```sh
sudo systemctl daemon-reload # reflect changes in systemd files
sudo systemctl enable cosmos-node-exporter # enable service autostart
sudo systemctl start cosmos-node-exporter # start a service
sudo systemctl status cosmos-node-exporter # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u cosmos-node-exporter -f --output cat
```

## How can I scrape data from it?

Here's the example of the Prometheus config you can use for scraping data:

```yaml
scrape-configs:
  - job_name: 'cosmos-node-exporter'
    scrape_interval: 10s
    static_configs:
      - targets: ['<your IP>:9500']
```

Then restart Prometheus and you're good to go!

## How does it work?

Well, here's the app schema:

![App schema](https://raw.githubusercontent.com/QuokkaStake/cosmos-node-exporter/main/assets/schema.png)

Sounds complex, huh? Let us explain.

We built this exporter to be as modular as possible, so it'd be easy to add new data fetching and new metrics.
Here's some terms we use within the app:
- `Fetcher` - an entity that fetches data from remote source (like RPC node); it may require some data from other fetchers
- `Controller` - an entity that fetches all the data from all provided Fetchers and generates a State
- `State` - an entity that represents an eventual result of all Fetchers execution
- `Generator` - an entity that generates some metrics based on State entity
- `NodeHandler` - an entity that fetches data and generates metrics for a specific node
- `App` - an entity that spawns a bunch of NodeHandlers per each chain, then assembles and returns metrics to a user

This allows to build complex schemas (like, we don't need to fetch block time to calculate time till upgrade
if there's no upgrade upcoming) and make it flexible and easy to add new Fetchers and Generators.

Fetchers can also be enabled/disabled, if a Fetcher is disabled, then it will provide no data
and therefore Generator that uses the data from that Fetcher won't provide any metrics.

Here's a list of Generators:

| Querier                     | Metrics returned                                                                                                                   | Per-node? | Requirements                                                                                 |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------------------|-----------|----------------------------------------------------------------------------------------------|
| AppVersionGenerator         | cosmos-node-exporter version                                                                                                       | No        |                                                                                              |
| UptimeGenerator             | App launch timestamp, useful for annotations                                                                                       | No        |                                                                                              |
| CosmovisorUpgradesGenerator | Whether the Cosmovisor binary is present for the upgrade                                                                           | Yes       | Cosmovisor config and the upcoming upgrade                                                   |
| CosmovisorVersionGenerator  | Cosmovisor version                                                                                                                 | Yes       | Cosmovisor config                                                                            |
| IsLatestGenerator           | Whether the local version is the same or greater than the latest GitHub/Gitopia release                                            | Yes       | Cosmovisor config (for local version), Git config (for fetching remote version)              |
| LocalVersionGenerator       | Local app binary version                                                                                                           | Yes       | Cosmovisor config                                                                            |
| NodeConfigGenerator         | Node's minimum-gas-prices and halt-height                                                                                          | Yes       | gRPC config, the chain should implement the `cosmos.base.node.v1beta1/Config` gRPC endpoint. |
| NodeInfoGenerator           | Running app version/git tag, cosmos-sdk version, Go version/build tags used to build it                                            | Yes       | gRPC config                                                                                  |
| NodeStatusGenerator         | Node's voting power, sync status, latest block time, node info, Tendermint/CometBFT version                                        | Yes       | Tendermint/CometBFT config                                                                   |
| RemoteVersionGenerator      | Latest release of this app published                                                                                               | Yes       | Git config (either Git or Gitopia)                                                           |
| TimeTillUpgradeGenerator    | Estimated upgrade time                                                                                                             | Yes       | Tendermint/CometBFT config (for fetching upgrade plan and block time)                        |
| UpgradesGenerator           | Upcoming upgrade info                                                                                                              | Yes       | Tendermint/CometBFT config                                                                   |

Additionally, per each Fetcher, the app will return the list of actions it did (like, querying a node, getting GitHub latest release etc.)
and whether they were successful a node. The exporter itself should never return an error or crash (if it does, please file an issue),
instead it will return all the data it could get, and additionally it'll return a metrics set with all the actions it could
or couldn't do. You can set alerts based on that, for example, if `node_status` action is failing for a big period of time,
likely the node is down.

All metrics are prefixed with `cosmos_node_exporter_`, to get the list of all metrics, try something like
`curl localhost:9500/metrics` on a fullnode the binary is running at and look at the results.


## How can I configure it?

All configuration is done via `.toml` config. Check config.example.toml for reference.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
