# Changelog

All notable changes to this project are documented in this file.

Release notes are generated from conventional commits by `mask release` and published to GitHub Releases from `CHANGELOG.md`.


## [0.4.0](https://github.com/niklas-heer/sceno/releases/tag/v0.4.0) (2026-07-19)

### Features
* **layout**: fit content to shape silhouettes ([fa93389](https://github.com/niklas-heer/sceno/commit/fa933898bd1291cb30c5a1340869b95a9f92273e))
* **render**: redraw clouds with silhouette-safe content ([415d982](https://github.com/niklas-heer/sceno/commit/415d982766dd1a24a05d14bbaeeead077930a59a))
* **render**: establish measurable visual composition ([637dee8](https://github.com/niklas-heer/sceno/commit/637dee8dff8f06317f69c46fe255c3edbba26b13))
* **feedback**: expose exact scene geometry and repairs ([5479f74](https://github.com/niklas-heer/sceno/commit/5479f7471ed3f06ea515fe20c46856a0e5099225))
* **engine**: add collision-safe hybrid layout controls ([7f026fa](https://github.com/niklas-heer/sceno/commit/7f026fa6345e63ee0c65c8ab26940cee0b496f60))
* **examples**: add architecture fixtures for visual regression ([cee749e](https://github.com/niklas-heer/sceno/commit/cee749e0a2f14c9baa157de0345e6c6f4522f4a7))
* **engine**: interior grid, anchor rules, and deterministic layout ([1f07869](https://github.com/niklas-heer/sceno/commit/1f07869c2a0935e4a2c7a8f98e8219fcce77cb67))

### Bug Fixes
* **tasks**: report dist version from build environment ([6a75ebb](https://github.com/niklas-heer/sceno/commit/6a75ebb1d705a5f25135b81fdbbc96d83465032d))
* **geom**: keep sketch arrow approaches straight ([907c30d](https://github.com/niklas-heer/sceno/commit/907c30dbef04b560c2a097d17f071e9b07691517))
* **layout**: collapse tiny connector jogs ([e0f41c5](https://github.com/niklas-heer/sceno/commit/e0f41c5404aaccb074957aaad8cddfaf96c727f0))
* **render**: align cylinder anchors and silhouettes ([bdce04a](https://github.com/niklas-heer/sceno/commit/bdce04a1c868d565a24fa7a23ddea1758b46f3c6))
* **render**: align actor and lane PDF semantics ([0505ddb](https://github.com/niklas-heer/sceno/commit/0505ddb76802680a07a924cc0fbe90ea8303ddbd))
* **geom**: keep edge labels clear of chrome ([43e8ef7](https://github.com/niklas-heer/sceno/commit/43e8ef70af7e1aa4ab2fe4a5f0fd023ff4a82b31))
* **layout**: fan out and polish connector routes ([bd29b53](https://github.com/niklas-heer/sceno/commit/bd29b53dd7661ccc6bf0bf54d3006aaa886b8185))
* **layout**: improve connector spacing and routing ([0d1be33](https://github.com/niklas-heer/sceno/commit/0d1be33023013b1518bef7b4e9ba3e4b946739ee))
* **feedback**: detect overlapping edge labels ([56b5847](https://github.com/niklas-heer/sceno/commit/56b584760b535c73ab7a956b315bad86a80c6ff8))
* **export**: preserve visual parity across formats ([9d6e1b2](https://github.com/niklas-heer/sceno/commit/9d6e1b2c2d910b8239f61ffe7c8e3a1e2470b1ce))
* **install**: fall back to ~/.local/bin when /usr/local/bin is not writable ([081b762](https://github.com/niklas-heer/sceno/commit/081b762f07dfab96dd46ad59de13c0e83fb4a1a8))

### Documentation
* **agent**: complete geometry feedback reference ([d4311b4](https://github.com/niklas-heer/sceno/commit/d4311b4b25d2244e3db32b90659defa10e99a7fb))
* **examples**: refresh README workflow render ([daf451d](https://github.com/niklas-heer/sceno/commit/daf451d769ac7c5e560832db0a16218774b4e7b3))
* **repo**: synchronize geometry and release references ([7233237](https://github.com/niklas-heer/sceno/commit/72332378a2c12fcf0a91d09cd2d18fe726b33909))
* **agent**: expose silhouette geometry in runtime guide ([0c14559](https://github.com/niklas-heer/sceno/commit/0c1455925f04fc041e08576bd9684ff355b5dd53))
* **render**: evaluate canvas backend consolidation ([4306cc4](https://github.com/niklas-heer/sceno/commit/4306cc49416e9a4398ddf3734499e9283d7e9340))
* **agents**: clarify visual feedback goals ([a3f898e](https://github.com/niklas-heer/sceno/commit/a3f898ef83ad49a762b3718c6370774dd43149ff))

### Tests
* **ci**: enforce visual quality across corpus ([4044ee3](https://github.com/niklas-heer/sceno/commit/4044ee3b8ffdb6cebead255699eae47442189881))
* **ci**: verify the complete KDL corpus ([1984aea](https://github.com/niklas-heer/sceno/commit/1984aea111f87d617cddda645a2536ac3f9f1509))
* add fuzz, property, mutation, and render audit suite ([fdf8020](https://github.com/niklas-heer/sceno/commit/fdf8020bfd7fbc855e648ad52dbaa12ca6993f51))

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

## [0.2.0](https://github.com/niklas-heer/sceno/releases/tag/v0.2.0) (2026-06-05)

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
