# stash-vr
Watch your [stash](https://github.com/stashapp/stash) library in VR.

Install stash-vr, point it to your stash instance and point your VR video player to stash-vr.\
stash-vr will relay your library information and display/play them in your video player ui.

## Supported video players
* HereSphere
* DeoVR

## Usage
Browse to `http://<host>:9666` using either DeoVR or HereSphere.
Both will automatically load their respective configuration files and launch their ui with your library.

## Installation
See [docker_compose.yml](docker-compose.yml) for details.

* `STASH_GRAPHQL_URL` Url to your stash graphql - something like `http://<stash.host>:<9999>/graphql`
* `STASH_API_KEY` Api key to you stash if it's using authentication. 

stash-vr listens on port `9666`, use docker port binding to change.

## Features
* Show following sections in DeoVR/HereSphere:
  - All (ALL your scenes)
  - Filters from your stash front page
  - Your other saved filters
* Provide transcoding endpoints to your videos served by stash

### HereSphere:
* Video tags from markers.
* Generate tags based on stash tags, studios, performers and markers.
  - Legend:
    - #:\<Tag>
    - Studio:\<Studio>
    - Performer:\<Performer>
    - @:\<Marker>
* Funscript support
* Ratings

### DeoVR
* Cue points from markers.

## 3D
Both DeoVR and HereSphere has algorithms to automatically detect and handle 3D videos.
It's not foolproof and to manually configure the players with custom layout/mesh-settings you can tag your scenes in stash as follows:

* Mesh:
  - `DOME` 180° equirectangular
  - `SPHERE` 360° equirectangular
  - `FISHEYE` 180° fisheye
  - `MKX200` 200° fisheye
  - `RF52` 190° Canon fisheye
  - `CUBEMAP` Cubemap (lacks support in DeoVR?)
  - `EAC` Equi-Angular Cubemap (lacks support in DeoVR?)
* Layout:
  - `SBS` Side-by-side (Default)
  - `TB` Top-bottom

If a mesh is provided but no layout then default layout `SBS` will be used.

Most common combination is DOME/SBS meaning most VR videos only need the `DOME` tag.

## Known issues/Missing features
* Premade Filters (i.e. Recently Released Scenes etc.) from stash front page are not supported.