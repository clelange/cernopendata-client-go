# Changelog

## [0.1.5](https://github.com/clelange/cernopendata-client-go/compare/v0.1.4...v0.1.5) (2026-01-18)


### Bug Fixes

* **build:** correct buildVersion package path in LDFLAGS ([44e7d85](https://github.com/clelange/cernopendata-client-go/commit/44e7d8537a3cb903c81b5f311987c674689c6221))
* **tests:** increase timeout in TestIntegrationListDirectoryTimeout to avoid hanging ([1fe1783](https://github.com/clelange/cernopendata-client-go/commit/1fe1783bede686c2f7f189fbfabdad3916c64c27))

## [0.1.4](https://github.com/clelange/cernopendata-client-go/compare/v0.1.3...v0.1.4) (2026-01-18)


### Bug Fixes

* download release assets instead of rebuilding for checksums ([dba9df7](https://github.com/clelange/cernopendata-client-go/commit/dba9df78d821a98bc422b24bea22c986f1619fae))

## [0.1.3](https://github.com/clelange/cernopendata-client-go/compare/v0.1.2...v0.1.3) (2026-01-18)


### Bug Fixes

* separate checksum upload into dedicated job that waits for binaries ([feacaf7](https://github.com/clelange/cernopendata-client-go/commit/feacaf79b8b7c24e3e9839ddff7374aedd5fe98a))

## [0.1.2](https://github.com/clelange/cernopendata-client-go/compare/v0.1.1...v0.1.2) (2026-01-18)


### Bug Fixes

* build and upload only specific binary per matrix job to avoid conflicts ([1ac3d33](https://github.com/clelange/cernopendata-client-go/commit/1ac3d33b3e75e3632a8415ce9a47205b893d25d1))

## [0.1.1](https://github.com/clelange/cernopendata-client-go/compare/v0.1.0...v0.1.1) (2026-01-18)


### Bug Fixes

* avoid race condition when uploading checksums by handling in single job ([09efbe8](https://github.com/clelange/cernopendata-client-go/commit/09efbe875c493e1368c3c131860d8763e9f90be1))

## 0.1.0 (2026-01-18)


### Features

* add path normalization and fix XRootD compatibility ([dc0243d](https://github.com/clelange/cernopendata-client-go/commit/dc0243d6381e0a2eef1b6fecdadd65c470e896f9))
* align download-files command flags with Python client ([0c33d86](https://github.com/clelange/cernopendata-client-go/commit/0c33d861f2c8abcb80aebd4da245b065c86de44a))
* improve CLI help text and add --no-expand flag to match Python client ([fa5d5c9](https://github.com/clelange/cernopendata-client-go/commit/fa5d5c9a9f2309e45f67c8d369bc6e059f0546ab))
* improve test coverage and Python CLI compatibility ([126ed6f](https://github.com/clelange/cernopendata-client-go/commit/126ed6fb3a643ccbee514a670f77b6cb755b3482))
* initial implementation of CERN OpenData client in Go ([7560e46](https://github.com/clelange/cernopendata-client-go/commit/7560e46fc771799fb4f0162a7b99e72768608573))
* update list-directory command to match Python client ([e0e19eb](https://github.com/clelange/cernopendata-client-go/commit/e0e19ebb2064241b079459d1867c7ebd862402bd))


### Bug Fixes

* handle XRootD EOF detection and add context cancellation ([8cde4e1](https://github.com/clelange/cernopendata-client-go/commit/8cde4e1a605a25c475b390d99898b65b99f1d013))
* rename to github.com/clelange/cernopendata-client-go ([13b4933](https://github.com/clelange/cernopendata-client-go/commit/13b4933965e9ed6760159d89bec15ad72175b1ce))


### Miscellaneous Chores

* release 0.1.0 ([2d1ea72](https://github.com/clelange/cernopendata-client-go/commit/2d1ea72b0fa182a52801955117686f650684269b))

## Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
