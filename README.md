# cosmos-node-exporter

![Latest release](https://img.shields.io/github/v/release/QuokkaStake/cosmos-node-exporter)
[![Actions Status](https://github.com/QuokkaStake/cosmos-node-exporter/workflows/test/badge.svg)](https://github.com/QuokkaStake/cosmos-node-exporter/actions)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-node-exporter.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-node-exporter?ref=badge_shield)

cosmos-node-exporter is a Prometheus scraper that scrapes some data to monitor your node, specifically you can set up alerting if:
- your app version does not match the latest on GitHub (can be useful to be notified on new releases)
- your voting power is 0 for a validator node
- your node is catching up
- there are chain upgrades your node does not have binaries for

## How can I set it up?

First, you need to download the latest release from [the releases page](https://github.com/QuokkaStake/cosmos-node-exporter/releases/). After that, you should unzip it, and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz <the filename you've just downloaded>
./cosmos-node-exporter <params>
```

To run it in detached mode in background, first, we have to copy the file to the system apps folder:

```sh
sudo cp ./cosmos-node-exporter /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/cosmos-node-exporter.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=Cosmos Exporter
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

If you're using cosmovisor, consider adding the same set of env variables as in your cosmovisor's systemd file, otherwise fetching app version would crash.

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

## What data can I get from it?

This exporter has multiple Queriers, each of them querying a node or external resource (like GitHub) in some way, then returns a set of metrics. Each Querier can be enabled or disabled based on the config. Here's the list of Queriers:

| Querier          | Metrics returned                                                                                                                   | Requirements                                                                                                                                                                                              |
|------------------|------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| NodeStatsQuerier | Voting power, node status<br>(catching up, time since latest block)                                                                | tendermint.address specified in config                                                                                                                                                                    |
| VersionsQuerier  | Local node version, remote node version,<br>whether the node is using the latest binary                                            | Cosmovisor config (for local config),<br>Github config (for remote version),<br>both (for checking if the version used is latest)                                                                         |
| UpgradesQuerier  | Whether there is an upcoming upgrade,<br>its data, estimated upgrade time and<br>whether the binary for the upgrade<br>is prepared | gRPC config (for getting the upgrade plan), Cosmovisor config (for getting the built binaries), Tendermint config (for getting<br>the upgrade time if the height upgrade is<br>specified for the upgrade) |

Additionally, each Querier returns the list of actions it did (like, querying a node, getting GitHub latest release etc.) and whether they were successful a node. The exporter itself should never return an error (if it does, please file an issue), instead it will return all the data it could get, and additionally it'll return a metrics set with all the actions it could or couldn't do. You can set alerts based on that, for example, if `node_status` action is failing for a big period of time, likely the node is down.

All metrics are prefixed with `cosmos_node_exporter_`, to get the list of all metrics, try something like `curl localhost:9500/metrics` on a fullnode the binary is running at and look at the results.

## How does it work?

It fetches some data from the local node by querying Tendermint RPC (listening on port 26657 by default), Cosmovisor binary and GitHub.

## How can I configure it?

All configuration is done via `.toml` config. Check config.example.toml for reference.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-node-exporter.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FQuokkaStake%2Fcosmos-node-exporter?ref=badge_large)