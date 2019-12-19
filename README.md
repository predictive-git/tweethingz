# tweethingz

One of the key features I'm not able to get from Twitter is daily follower histogram. Sure there are all kinds of  on-line services that provide that feature but having used two of them only to have it either shut down or start asking way too much money for that service I decide to write my own.

![](static/img/sign-in.png)

## Overview

When I decided to build `tweethingz` service I had three objectives:

1. Multiple Twitter accounts support
2. Automatic (i.e. no based on me accessing it every day)
3. Serverless, specifically based on per-use charge model

Thee resulting Cloud Run based service with Cloud Firestore data persistence and Cloud Scheduler service execution "cron" are the three core GCP services with which I chose to build `tweethingz`.

## One-click Setup

To deploy `tweethingz` into your GCP project just click on the below button and follow the prompts.

![](static/img/sign-in.png)

## Manual Setup

To deploy `tweethingz` manually with an opportunity to customize it follow these steps:

### Configuration

### Setup Dependencies

### Build Image

### Deploy Service

### Schedule Refresh Worker

## Cleanup

To cleanup all resources created by `tweethingz` execute

```shell
bin/cleanup
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.


