.PHONY: build

build:
	goreleaser release --snapshot --skip-publish --rm-dist

release-test:
	goreleaser release --skip-publish --rm-dist

release:
	goreleaser release --rm-dist
	rsync -a build/ /keybase/public/justmiles/artifacts/ssm-run-command/