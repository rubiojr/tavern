# Tavern

A few serve, the rest enjoy the free drinks.

## ⚠️ Work in progress, experimental

Tavern currently needs a bleeding edge (unreleased) Charm server (due to https://github.com/rubiojr/tavern/pull/6).

As the Tavern client downloads, decrypts and publishes to the world files available in CharmFS, it's highly recommended you setup your own Charm server locally to test Tavern, or use it with a Charm test account where you don't have sensible files that can't be published, until the first release is published. You [can also setup your own Tavern server](#hosting-your-own-tavern-server) so instead of using https://pub.rbel.co.

## Overview

Tavern is a command line tool to publish static files available in the [Charm cloud](https://charm.sh) to a Tavern server where the files will be publicly available.

When publishing, documents, images, [static websites](https://gohugo.io) or any other file you have in CharmFS under the directory specified as a `tavern publish` argument (works with individual files also), will be downloaded, decrypted and published to a Tavern server (defaults to https://pub.rbel.co).

## Usage

To use Tavern (both client and server), you'll need a [Charm Cloud](https://charm.sh/cloud) account already setup.

The Tavern client uses CharmFS to download the files to publish and the Tavern server uses Charm accounts to allow you to publish files. Each Tavern user gets a directory in a Tavern server, and that directory is named after your Charm ID.

## Security

Tavern is **experimental**, using it with charm accounts where you have valuable data or publishing to a public Tavern server is discouraged until the first official release is published.

When the Tavern client publishes files (see [Publishing](#publishing)), it:

* Requests a JWT token from Charm (cloud or your own charm server) with a `tavern` audience
* Adds the following HTTP headers to the request that will be sent to the Tavern server:
  * The JWT token received from Charm as the `Authorization` header
* Sends a POST to the Tavern server with the headers and a multipart form with the files downloaded from CharmFS

The Tavern server will:

* [Validate the JWT](https://auth0.com/blog/navigating-rs256-and-jwks) token and optionally the issuer (Charm server), if `--allowed-charm-servers` is specified
* Allow you to publish the files if the JWT is valid and the source Charm server is allowed
* Write the files to its local files system, under `tavern_uploads/<your-Charm-ID>`.

### Publishing

The Tavern client will publish to https://pub.rbel.co by default, where I host a testing Tavern server. To use your own Tavern server, use `--server-url` with `publish` or export `TAVERN_SERVER_URL`.

_Note: Please use the public test server for testing only, it'll go down several times during the day while I improve Tavern, and content can be removed regularly._

To publish a directory available in Charm:

```
tavern publish /site/public

Publishing to /
Adding  /404.html
Adding  /index.html
Adding  /sitemap.xml
Adding  /tags/index.html
Adding  /tags/index.xml
Adding  /index.xml
Adding  /categories/index.html
Adding  /categories/index.xml
Adding  /page/1/index.html
Adding  /ananke/css/main.min.css
Adding  /images/gohugo-default-sample-hero-image.jpg
Adding  /assets/css/stylesheet.min.c88963fe2d79462000fd0fb1b3737783c32855d340583e4523343f8735c787f0.css
Site published!
Visit https://pub.rbel.co/216c5634-9d63-48de-9106-bfd04483aa00
```

This will publish everything under `charm:/site/public` to `https://pub.rbel.co/<your-charm-id>`. **Note:** this makes private CharmFS files available to the rest of the Internet population, make sure you only publish files that can be public!.

A sample script I use to publish [my website](https://me.rbel.co), that I have hosted in my own charm server:

```sh
#!/bin/sh
set -e

echo "Building the site..."
cd ~/Documents/site && hugo
echo "Updating CharmFS site..."
charm fs cp -r public charm:site
echo "Publishing to Tavern..."
tavern publish site/public
```

If you want to publish files in your own Charm server:

```
# use --charm-server-http-port and --charm-server-ssh-port if you are using non-standard Charm ports
tavern publish --charm-server-host your.charm.server site/public
```

### Hosting your own Tavern server

```
tavern serve
2021/12/21 14:14:09 serving on 0.0.0.0:8000
2021/12/21 14:14:09 uploads directory: tavern_uploads
```

Note that this allows anyone with a Charm account **in any Charm server** to publish files to this Tavern server.

Or using docker:

```
docker run ghcr.io/rubiojr/tavern:latest
```

You'll need to export `TAVERN_SERVER_URL` environment variable or use the tavern client with `--server-url`:

```
tavern publish --server-url https://my-tavern-server.com /site
```

#### Running Tavern against your own Charm server

If you want Tavern's server to auth against your own Charm server (and only that one):

```
tavern serve --allowed-charm-servers your.charm.server
```

You'll also need to use `--charm-server-host` when publishing with Tavern client:

```
tavern publish --charm-server-host your.charm.server site/public
```