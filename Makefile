MESHCAT_REF ?= main

.PHONY: refresh-viewer-assets
refresh-viewer-assets:
	go run ./scripts/refresh_viewer_assets -ref $(MESHCAT_REF)
