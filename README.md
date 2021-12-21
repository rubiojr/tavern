# Tavern

A few serve, the rest enjoy the free drinks.

**⚠️ Work in progress, experimental**

Tavern is a command line tool to publish static files available in the [Charm cloud][https://charm.sh) to a Tavern server where the files will be publicly available.

Documents, images, [static websites](https://gohugo.io) or any other file you have in CharmFS will be downloaded, decrypted and published to a Tavern server.


## Usage

### Publishing

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

This will publish everything under `charm:/site/public` to https://pub.rbel.co/<your-charm-id>`.

A sample script I use to publish [my website](https://hello.rbel.co), that I have hosted in my own charm server:

```
#!/bin/sh
echo "Updating CharmFS site..."
charm fs cp -r ~/Documents/site charm:
echo "Publishing to Tavern..."
tavern publish site/public
```

### Hosting your own Tavern server

```
tavern serve
2021/12/21 14:14:09 serving on 0.0.0.0:8000
2021/12/21 14:14:09 uploads directory: tavern_uploads
```

Or using docker:

```
docker run ghcr.io/rubiojr/tavern:latest

You'll need to export `TAVERN_SERVER_URL` environment variable or use the tavern client with `--server-url`:

```
tavern publish --server-url https://my-tavern-server.com /site
```