# tweethingz

One of the key features I'm not able to get from Twitter is a daily follower histogram. Sure there are on-line services that provide that feature but having used two of them already only to watch each one eventually shut down or price me out I decide to write my own.

![](doc/dashboard-histogram.png)

## Overview

I decided to build `tweethingz` with three "simple" objectives:

1. Multi-tenant (i.e. support for multiple Twitter account)
2. Automatic (i.e. daily histogram built without me needing to check)
3. Serverless (specifically per-usage charge model, and zero compute, scheduling, and data infra to manage)

The resulting `tweethingz` is built on Google's Cloud Run with Cloud Firestore data persistence and Cloud Scheduler execution ("cron") are the three core GCP services with which I chose to build `tweethingz`.

## Setup

To deploy `tweethingz` you will need first clone this repo:

```shell
git clone https://github.com/mchmarny/tweethingz.git
cd tweethingz
```

Once you clone the `tweethingz` repo you can follow the the following steps:

### Configuration

The `tweethingz` service exposes a few configuration values but only 3 are required to edit. First, the Twitter API consumer keys (`TW_KEY` and `TW_SECRET`). You can you set up your app credentials [here](https://developer.twitter.com/en/apps/create).

Additionally, you can lock down `tweethingz` service to either list of specific users or allow all new users to register through this service using `TW_USERS`  Don't worry, they will be using their own API keys once they register.

> Note, leaving the `TW_USERS` value undefined will allow new users to register, while defining any number of twitter username (separated by space), will cause any Twitter authenticated users not defined to be rejected.

Both, the API keys, the user lists,  as well as a few other configuration values are defined in the [bin/config](bin/config):

```shell
export TW_KEY="" #Your Twitter API consumer key
export TW_SECRET="" #Your Twitter API consumer secret
export TW_USERS=""
```

### Setup Dependencies

Now that the configuration values are defined, you can setup the service dependencies:

```shell
bin/setup
```

The `setup` script will:

* Enable the necessary GCP APIs (e.g. run, firestore, scheduler etc.)
* Create a secure token for Scheduler to invoke Run service
* Create service account in IAM under which the Run and Scheduler will be running
* and grant that service account the minimum number of the necessary roles (e.g. `run.invoker`)

> As with any script, you should review its content before executing.

### Build Image

Now that the dependencies are set up, you can build the Docker image which will be used to deploy Cloud Run service:

```shell
bin/image
```

If everything goes well you will see something similar to this in your console:

```shell
ID            CREATE_TIME          DURATION  SOURCE                      IMAGES                       STATUS
610bfc9b-...  2019-12-31T22:34:31  1M33S     gs://...cloudbuild/source   gcr.io/.../tweethingz:0.4.9  SUCCESS
```

> Note, if you are familiar with Docker you can build this image locally but you will have to publish it to GCR before it can be deployed into Cloud Run.

### Deploy Service

With the image built, you are now ready to deploy the service to Cloud Run:

```shell
bin/deploy
```

The above script will deploy the previously built image to Cloud Run along with all the necessary configuration and IAM policy bindings. If everything goes OK, you should see response similar to this in your console:

```shell
Deploying container to Cloud Run service [tweethingz] in project [...] region [us-central1]
✓ Deploying... Done.
  ✓ Creating Revision...
  ✓ Routing traffic...
  ✓ Setting IAM Policy...
Done.

Service [tweethingz] revision [tweethingz-00001-man] has been deployed and is serving 100 percent of traffic at https://tweethingz-...-uc.a.run.app
Updated IAM policy for service [tweethingz].
bindings:
- members:
  - allUsers
  - serviceAccount:tweethingz@....iam.gserviceaccount.com
  role: roles/run.invoker
version: 1
```

### Schedule Refresh Worker

The final step in configuring `tweethingz` is to set up Cloud Scheduler so that even when you don't access the Cloud Run service some days your follower histogram will be kept up to date:

```shell
bin/schedule
```

The default frequency for the `tweethingz` update is 30 min. You can change that if you need to in the [bin/config](bin/config) file.

## Usage

The `tweethingz` is pretty much self-explanatory but here are few short steps to guid you

### Authorization

Before you will be able to access Twitter you will have to authorize `tweethingz` to invoke the API on your behalf. Just clock on the "Sign in with Twitter" button on the home page and follow the guide.

> Note, the `tweethingz` service requires only read-only access to your Twitter profile. It uses only the data already available publicly to any of your followers and is unable to post on yur behalf.

![](doc/twitter-auth.png)

## Cleanup

To cleanup all resources created by `tweethingz` execute

```shell
bin/cleanup
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.


