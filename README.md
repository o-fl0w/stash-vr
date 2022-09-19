# stash-vr
Watch your [stash](https://github.com/stashapp/stash) library in VR.

[Install](#Installation) stash-vr, point it to your stash instance and point your VR video player to stash-vr.\
stash-vr will relay your library information and display/play them in your video player ui.

## Supported video players
* HereSphere
* DeoVR

## Usage
Browse to `http://<host>:9666` using either DeoVR or HereSphere.
Both will automatically load their respective configuration files and launch their ui with your library.

## Installation
See [docker_compose.yml](docker-compose.yml) for details.

* `STASH_GRAPHQL_URL` 
  * **Required**
  * Url to your stash graphql - something like `http://<stash.host>:<9999>/graphql`
* `STASH_API_KEY` 
  * Api key to your stash if it's using authentication. 
* `FAVORITE_TAG` 
  * Name of tag in stash to hold scenes marked as [favorites](#favorites) (will be created if not existing)
  * Default: `FAVORITE`
* `FILTERS`
  * Select the filters to show by setting one of below values
    * `frontpage`
      * Only filters found on stash front page
    * Comma separated list of filter ids, i.e. `1,5,12`
      * Only filters with provided filter ids
    * Empty
      * All saved filters
  * Default: Empty 
* `HERESPHERE_SYNC_MARKERS`
  * Enable sync of Marker from HereSphere [NOTE](#heresphere-sync-of-markers)
  * Default: `false`
* `HERESPHERE_QUICK_MARKERS`
  * HereSphere displays all tags on track 0 above the seekbar. By default, stash-vr puts studio and tags on track 0 for context at a quick glance. If this is set to `true` stash-vr will for quick access instead put Markers on track 0 if they exist. 
  * Default: `false`

stash-vr listens on port `9666`, use docker port binding to change.

## Features
* Show following sections in video player:
  - Filters from your stash front page
  - Your other saved filters
* Provide transcoding endpoints to your videos served by stash

### HereSphere:
* Two-way sync
  * Rating
  * Tags
  * Studio
  * Performers
  * Markers
  * Favorites
* Generate categorized tags
* Delete scenes
* Funscript
* O-counter incrementing

### DeoVR
* Markers

## Usage
### HereSphere
##### Two-way sync
To enable two-way sync with stash all toggles (`Overwrite tags` etc.) needs to be on, in the cogwheel at the bottom right of preview view in HereSphere.
#### Manage metadata
Video metadata is handled using `Video Tags`.

To tag a video open it in HereSphere and click `Video Tags` above the seekbar.
On any track insert a new tag and prefix it with `#:` i.e. `#:MusicVideo`.
This will create the tag/studio/performer `MusicVideo` in stash if not already present and apply it to your video.

Same workflow goes for setting studio and performers but with different prefixes according to below:

|Metadata|Prefix| Alias |
|--------|------|-------|
|Tags|`#:`|`Tag:`|
|Studio|`$:`|`Studio:`|
|Performers|`@:`|`Performer`|

#### Markers
(Both stash and HereSphere use the word _tag_ but they use it differently. Tags in heresphere are akin to Markers in stash)

Markers in stash need a primary tag. Marker title is optional.
To create a marker using HereSphere open your video and create a "tag" on any track using `Video Tags`.
The naming format is:
* `<tag>:<title>` will create a Marker in stash titled `<title>` with the primary tag `<tag>`
* `<tag>` will create a Marker in stash with primary tag `<tag>` and no title.
  
Set the start time using HereSphere controls. 
Tags (markers) in HereSphere has support for both a start and end time. 
Stash currently defines Markers as having a start time only. This means the end time set in HereSphere will be ignored.

#### Favorites
When the favorite-feature of HereSphere is first used stash-vr will create a tag in stash named according to `FAVORITE_TAG` (set in docker env., defaults to `FAVORITE`) and apply that tag to your video.

**Tip:** Create a filter using that tag, so it shows up in HereSphere for quick access to favorites.

#### Rating
HereSphere uses fractions for ratings, i.e. 4.5 is a valid rating. Stash uses whole number.
If you set a half star in HereSphere stash-vr will round up the rating. That is if you set a rating of 3.5 the video will receive a rating of 4 in stash.
In other words, click anywhere on a star to set the rating to that amount of stars.

**Exception:** To remove a rating, rate the video 0.5 (half a star). 

#### O-counter
Add a tag with the name `!O` (case-insensitive) in `Video Tags` to increment the o-counter.

## VR
Both DeoVR and HereSphere has algorithms to automatically detect and handle VR videos.
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

Most common combination is `DOME`+`SBS` meaning most VR videos only need the `DOME` tag.

## Known issues/Missing features
* DeoVR/HereSphere both seem to have limits and struggle/crash when too many videos are provided than they can handle.
  * For HereSphere the limit seems to be around 10k unique videos.
  * Tip: If you have a VERY LARGE library and your player is struggling to load them all, try explicitly setting env. var. `FILTERS` with a list of filter ids such that the total amount of videos are lowered to a "reasonable" amount.

#### Unsupported filter types
* Premade Filters (i.e. Recently Released Scenes etc.) from stash front page are not supported.
* Any other filter type besides scene filters
#### HereSphere sync of Markers
When using `Video Tags` in HereSphere to edit Markers stash-vr will delete and (re)create them on updates.
There currently is no support for correlating the markers (tags) in HereSphere to a Marker in stash.
This means that **all metadata, besides the primary tag and title, related to a marker will NOT be retained** (id, previews, secondary tags and created/updated time). If you're not using those fields anyway you probably won't notice the difference.

#### Reflecting changes made in stash
When the index page of stash-vr is loaded stash-vr will immediately respond with a cached version. At the same time stash-vr will request the latest data and store it in the cache for the next request.
This means if changes are made in stash and the player refreshed, it will receive the cached version built during the last (previous) request.
Just refresh again and the player should receive the latest changes. In other words, refresh twice.
