# stash-vr
Self-hosted service for watching your [stash](https://github.com/stashapp/stash) library in VR.

Basically a thin client/server for clients listed below.

## Supported clients
* HereSphere
* DeoVR

## Usage
Browse to `http://<host>:9666` using either DeoVR or HereSphere.
Both will automatically load their respective configuration files and launch their ui with your library.

## Installation
See [docker_compose.yml](docker-compose.yml) for details.

Provide the url to your stash graphql through environment variable `STASH_GRAPHQL_URL`, something like `http://<stash.host>:<9999>/graphql`.

stash-vr listens on port `9666`, user docker port binding to change.

## Features
* Show following sections in DeoVR/HereSphere:
  - All (ALL your scenes)
  - Sections from your front page
  - Saved searches

* Provide transcoding endpoints to your videos served by stash

### HereSphere:
* Tags for browsing/filtering.
* Video tags from markers.
* Legend:
  - #:\<Tag>
  - Studio:\<Studio>
  - Performer:\<Performer>
  - @:\<Marker>
* Funscript support
* Ratings

### DeoVR
* Cue points from markers.