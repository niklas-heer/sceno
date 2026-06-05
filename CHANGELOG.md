# Changelog

All notable changes to this project are documented in this file.

Release notes are generated from conventional commits by `mask release` and published to GitHub Releases from `CHANGELOG.md`.


## [0.3.0](https://github.com/niklas-heer/sceno/releases/tag/v0.3.0) (2026-06-05)

### Features
* **icons**: add catalog with CLI docs and validation rules ([ed7b7e2](https://github.com/niklas-heer/sceno/commit/ed7b7e2c4b8a53c3da9a5f811220061fb363f009))
* **pipeline**: unify build and scene evaluation across CLI ([a588d10](https://github.com/niklas-heer/sceno/commit/a588d109b30cf979da27535f27d951b478ce039c))
* **docs**: add goals JSON and align validation agent workflow ([9ae253c](https://github.com/niklas-heer/sceno/commit/9ae253c24a2eecd830cba8ae53228a058bfd0c06))
* **validate**: add stack engine and advise command ([9581414](https://github.com/niklas-heer/sceno/commit/9581414de3d529d56619c0d8df17d66b51c9251a))
* **layout**: improve alignment, icons, and edge labels ([ad53b74](https://github.com/niklas-heer/sceno/commit/ad53b74077ea4fbe79b29ade2eda912613f2cb88))
* **release**: generate release notes from conventional commits ([a065f10](https://github.com/niklas-heer/sceno/commit/a065f10a9bbc7dbcfa0eb1def872f5b293f80cbc))

### Bug Fixes
* **render**: land arrow tips on borders with continuous labeled connectors ([9376728](https://github.com/niklas-heer/sceno/commit/9376728fc905111b847c66c40d763c58348e6244))
* **render**: place edge labels above nodes with connector gaps ([3fffc6d](https://github.com/niklas-heer/sceno/commit/3fffc6d4452b49b3c79ab7688ec0edea28f3862b))
* **scene**: enforce edge label clearance and border anchoring rules ([06d7af3](https://github.com/niklas-heer/sceno/commit/06d7af3057ad7b2cb898fbdeb811e3f6e01a6473))
* **export**: correct PNG icons, PDF fonts, and multi-slide --all ([5db70d9](https://github.com/niklas-heer/sceno/commit/5db70d93347d17ea22087bc1a69a2f8fcf09f2ff))

### Refactoring
* **cli**: simplify to seven core commands with legacy aliases ([67cd7f3](https://github.com/niklas-heer/sceno/commit/67cd7f380f88228054060264e71d43ea08e45919))
* **docs**: generate documentation from code at runtime ([4926d56](https://github.com/niklas-heer/sceno/commit/4926d56f063db295db4e8789ae891240df32da53))

### Documentation
* **examples**: refresh README diagram and add visual audit fixtures ([0459a20](https://github.com/niklas-heer/sceno/commit/0459a208cba1567bf2a8f37f2730c25def8a6f44))
* sync references to docs hub and refresh README diagram ([eaf34e1](https://github.com/niklas-heer/sceno/commit/eaf34e1d9638c2058cd3ba45e31d219730a172d9))
* refresh README workflow diagram with advise and describe steps ([610b0dd](https://github.com/niklas-heer/sceno/commit/610b0ddbb847039191386a43fc956c8723ac44bd))
* **examples**: polish README diagram with advise and iconPos ([58c7bd8](https://github.com/niklas-heer/sceno/commit/58c7bd897f9c79e36de5cd6a73357ffb9d4ed188))
* document stack validation, advise, and self-doc topics ([8c60620](https://github.com/niklas-heer/sceno/commit/8c60620156ebb2a63270d1d78cd7a1b056cca36c))
* add dogfooded how-it-works diagram to README ([fdb501a](https://github.com/niklas-heer/sceno/commit/fdb501afc81039a6c402fcd0439cb93e4b56cb0b))
* **install**: default to latest release and update examples ([3c1609c](https://github.com/niklas-heer/sceno/commit/3c1609cba19cbfa68e20bc01a52be7512b0d0984))

## [0.2.0](https://github.com/ssh://git@github.com/niklas-heer/sceno/releases/tag/v0.2.0) (2026-06-05)

### Features
* add automated semver release scripts
* migrate pipeline to Go + Dagger

### Bug Fixes
* pin dagger-for-github action to v8.4.1 in release workflow
* pin dagger-for-github action to v8.4.0
* pin dagger-for-github action to v8.4.1

### Other
* refactor(tasks): replace Makefile with mask task runner
## [0.1.0](https://github.com/niklas-heer/sceno/releases/tag/v0.1.0) (2026-06-05)

### Features

* Initial public release — KDL diagrams, Dagger CI, docs command, edge labels
