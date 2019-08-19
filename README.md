# ebs-snapshot-exporter

## Overview

Export AWS EBS Snapshot metrics in Prometheus format.

## Features

- Ability to add ad-hoc labels in the form of AWS Tags to the exported metrics.
- Filter EBS Snapshots using standard AWS Filters.
- Ability to register multiple exporter in form of Jobs to query multiple regions and AWS Accounts.
- Support for `Assume Role` while authenticating to AWS using Role ARN.

## Table of Contents

- [Getting Started](#getting-started)
  - [How it Works](#how-it-works)
  - [Installation](#installation)
  - [Quickstart](#quickstart)
  - [Sending a sample scrape request](#testing-a-sample-alert)

- [Advanced Section](#advanced-section)
  - [Configuration options](#configuation-options)
  - [Setting up Prometheus](#setting-up-prometheus)

## Getting Started

### How it Works

`ebs-snapshot-exporter` uses [AWS SDK](https://github.com/aws/aws-sdk-go) to authenticate with AWS API
and fetch Snapshots metdata. You can specify multiple `jobs` to fetch EBS Snapshots data and this exporter will collect all metrics and export in the form of Prometheus metrics.

You will need an _IAM User/Role_ with the following policy attached to the server from where you are running this program:

```plain
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeSnapshotAttribute",
                "ec2:DescribeSnapshots",
                "ec2:DescribeImportSnapshotTasks"
            ],
            "Resource": "*"
        }
    ]
}
```

### Installation

There are multiple ways of installing `ebs-snapshot-exporter`.

### Running as docker container

[mrkaran/ebs-snapshot-exporter](https://hub.docker.com/r/mrkaran/ebs-snapshot-exporter)

`docker run -p 9608:9608 -v /etc/ebs-snapshot-exporter/config.toml:/etc/ebs-snapshot-exporter/config.toml mrkaran/ebs-snapshot-exporter:latest`

### Precompiled binaries

Precompiled binaries for released versions are available in the [_Releases_ section](https://github.com/mr-karan/ebs-snapshot-exporter/releases/).

### Compiling the binary

You can checkout the source code and build manually:

```bash
git clone https://github.com/mr-karan/ebs-snapshot-exporter.git
cd ebs-snapshot-exporter
make build
cp config.sample config.toml
./ebs-snapshot-exporter
```

### Quickstart

```sh
mkdir ebs-snapshot-exporter && cd ebs-snapshot-exporter/ # copy the binary and config.sample in this folder
cp config.toml.sample config.toml # change the settings like server address, job metadata, aws credentials etc.
./ebs-snapshot-exporter # this command starts a web server and is ready to collect metrics from EC2.
```

### Testing a sample scrape request

You can send a `GET` request to `/metrics` and see the following metrics in Prometheus format:

```bash
# HELP ebs_snapshots_start_time Start Timestamp of EBS Snapshot
# TYPE ebs_snapshots_start_time gauge
ebs_snapshots_start_time{job="public",progress="100%",region="ap-south-1",snapshot_id="redacted",state="completed",vol_id="redacted",service="redacted"} 1.562355284e+09
# HELP ebs_snapshots_up Could the AWS EC2 API be reached.
# TYPE ebs_snapshots_up gauge
ebs_snapshots_up 1
# HELP ebs_snapshots_version Version of ebs-snapshot-exporter
# TYPE ebs_snapshots_version gauge
ebs_snapshots_version{build="5161e83 (2019-07-09 15:35:59 +0530)"} 1
# HELP ebs_snapshots_volume_size Size of volume assosicated with the EBS snapshot
# TYPE ebs_snapshots_volume_size gauge
ebs_snapshots_volume_size{job="public",progress="100%",region="ap-south-1",snapshot_id="redacted",state="completed",vol_id="redacted",service="redacted"} 50
```

## Advanced Section

### Configuration Options

- **[server]**
  - **address**: Port which the server listens to. Default is *9608*
  - **name**: _Optional_, human identifier for the server.
  - **read_timeout**: Duration (in milliseconds) for the request body to be fully read) Read this [blog](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) for more info.
  - **write_timeout**: Duration (in milliseconds) for the response body to be written.

- **[app]**
  - **log_level**: "production" for all `INFO` level logs. If you want to enable verbose logging use "debug".
  - **jobs**
    - **name**: Unique identifier for the job.
    - **exported_tags**: List of EC2 Tags which are available as labels in the metrics exported. If you have any custom EC2 Tags that you want to scrape with other labels, you can add it here.
    - **filters**:
      - **name**: Name of the AWS Filter key.
      - **value**: Value of the Filter key. Read more about [Filter API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html).
    - **aws_creds**:
      - **region**: AWS Region where your snapshots are hosted.
      - **access_key**: AWS Access Key if you are using an IAM User. It overrides the env variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.
      - **secret_key**: AWS Secret Key. (See above)
      - **role_arn**: Role ARN if you want to `assume` another role from your IAM Role. This is particularly helpful to scrape data across multiple AWS Accounts.

**NOTE**: You can use `--config` flag to supply a custom config file path while running `ebs-snapshot-exporter`.

### Setting up Prometheus

You can add the following config under `scrape_configs` in Prometheus' configuration.

```yaml
  - job_name: 'ebs-snapshots'
    metrics_path: '/metrics'
    static_configs:
    - targets: ['localhost:9608']
      labels:
        service: ebs-snapshots
```

Validate your setup by querying `ebs_snapshots_up` to check if ebs-snapshot-exporter is discovered by Prometheus:

```plain
`ebs_snapshots_up{instance="localhost:9608",job="ebs-snapshots",service="ebs-snapshots"} 1`
```

### Example Queries

- Count of EBS Snapshots: `count(ebs_snapshots_start_time{mytag="somethingcool"})`
- Last EBS Snapshot age in hours: `(min(time()-ebs_snapshots_start_time{exported_job="myjob"}) by (service) / 3600)`
- Last unsuccesful snapshot age in hours: `(min(time()-ebs_snapshots_start_time{state!="completed"}) by (service) / 3600)`
- Volume size of EBS for which snapshot is taken: `ebs_snapshots_volume_size{mytag="somethingcool"}`

### Example Alerts

<details><summary>Alert when no snapshot is taken in last 3 hours</summary><br><pre>
- alert: EBSSnapshotFailed
  expr: ebs:last_failed_snapshot_age_in_hours >= 3
  for: 1m
  labels:
    room: production-alerts
    severity: warning
  annotations:
    description: EBS Snapshots seems to be not working for service {{ $labels.service }}.
    title: EBS Snapshot failed.
    summary: Please check the AWS DLM lifecycle policy and rules.
</pre></details>


## Contribution

PRs on Feature Requests, Bug fixes are welcome. Feel free to open an issue and have a discussion first. Contributions on more alert scenarios, more metrics are also welcome and encouraged.

Read [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## License

[MIT](license)
