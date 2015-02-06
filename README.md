# Satellite - A remote server for Grow-powered sites.

[![Build Status](https://travis-ci.org/grow/satellite.svg)](https://travis-ci.org/grow/satellite)

A satellite server is a static web server for web pages produced by Grow. Basically: Do your development in Grow. Serve your site via satellite.

Satellite offers a few benefits over serving your site directly from GCS or S3, including:

* Authentication: Restrict access by password protecting your site
* Multi-site hosting: Manage the website for multiple domains using the same app instance
* (Soon) Staging: Stage and preview changes to your site to be published at a later date
* (Soon) Experiments: Run A/B tests like a boss
* (Soon) Analytics: Keep track of the key performance metrics that matter to you

Powered by Go. Runs on App Engine.

**NOTE:** Satellite is still under development. Not yet ready for general public use.

## Currently supported features

### Google Cloud Storage

Serve files directly from GCS. Publish your Grow files directly to GCS and the site will automatically be updated.

### Authentication

Satellite currently supports user authentication via HTTP Basic Authentication. Currently, enabling basic auth will password-protect your entire site (in the future, we will allow for more fine-grained access restriction).
