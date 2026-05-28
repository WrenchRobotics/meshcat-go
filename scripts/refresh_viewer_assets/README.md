# refresh_viewer_assets

This folder contains a small maintainer tool for updating vendored MeshCat viewer files in `viewer_assets/dist`.

It clones upstream MeshCat into a temporary directory, so a local git submodule checkout is not required.

## What it does

- Clones upstream MeshCat
- Checks out a specific ref (branch, tag, or commit)
- Runs `npm install` and `npm run build`
- Replaces local `viewer_assets/dist` with the built upstream `dist`
- Prints the resolved upstream commit SHA

## Usage

From the repository root:

```bash
go run ./scripts/refresh_viewer_assets -ref main
```

You can use any valid git ref:

```bash
go run ./scripts/refresh_viewer_assets -ref v0.0.1
go run ./scripts/refresh_viewer_assets -ref 65781fcb064db536b99a66fe9fcf5bf0b6d1f790
```

Optional flags:

- `-repo-url` to change the upstream MeshCat remote
- `-out` to change the destination directory

## Requirements

- `git`
- `npm`
- Go toolchain
