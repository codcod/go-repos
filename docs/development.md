## Development

```sh
go mod tidy
go build -o repos ./cmd/repos
```

### New release

Releases are done automatically and follow semantic versioning. Prior to this
automatization the following commands were used to release a version:

```sh
git tag v1.0.x
git push origin v1.0.x

# To create a release go to github.com or use gh tool:
gh release create v1.0.x --title "Repos v1.0.x" --notes "Release notes for version 1.0.x"
```