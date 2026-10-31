# Releases

`UPSTREAM` pins the commit of a Box2D release tag, with the tag as a comment. Do not pin untagged upstream commits.

Versions are `v0.x.y`:

- Minor: a new upstream release, or a Go API break.
- Patch: Go-only fixes on the same upstream release.

The first line of each release note names the ported Box2D version and commit, for example `Ports Box2D v3.2.0 (178495b)`.

## Porting an upstream release

1. Set `UPSTREAM` to the tag's commit and run `just oracle`.
2. Port `git diff <old>..<new> -- src include` and update the version in `box2d/base.go`.
3. Run the checks in `AGENTS.md`, plus `just oracle-check` and `just stress`.
4. Tag `vX.Y.0` and publish the release note.
