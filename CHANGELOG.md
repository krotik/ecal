# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [1.6.1](https://devt.de///compare/v1.6.0...v1.6.1) (2021-05-03)


### Bug Fixes

* Adding locking for ThreadPool.State() ([71c8b6d](https://devt.de///commit/71c8b6df3d56b0a15dcb601d7a9f8dd96f936d79))

## [1.6.0](https://devt.de///compare/v1.5.2...v1.6.0) (2021-04-18)


### Features

* Adding rand function ([6f8b42c](https://devt.de///commit/6f8b42c21fa28526f91ff0b0eac449688052bf93))


### Bug Fixes

* Minor pretty printer changes and addition to the documentation ([cbaba21](https://devt.de///commit/cbaba21419ea8d24393574aeda2f8394e92e60ba))

### [1.5.2](https://devt.de///compare/v1.5.1...v1.5.2) (2021-03-27)


### Bug Fixes

* Pretty printer fix for loop statements using an in operator ([643790e](https://devt.de///commit/643790ec49cde2db5a37fde329ef438a805e85e7))

### [1.5.1](https://devt.de///compare/v1.5.0...v1.5.1) (2021-03-27)


### Bug Fixes

* Exposing threadpool a bit more ([79a2d93](https://devt.de///commit/79a2d932589b02b6d8d229374340d8772c313593))

## [1.5.0](https://devt.de///compare/v1.4.6...v1.5.0) (2021-03-27)


### Features

* Adding profiling command ([972b895](https://devt.de///commit/972b895abc2369e7cf9429d35f6642952a1f1df8))


### Bug Fixes

* Better mutex management ([80efb81](https://devt.de///commit/80efb818410aef03d5313bbfd16b305230ce958c))

### [1.4.6](https://devt.de///compare/v1.4.5...v1.4.6) (2021-01-31)


### Bug Fixes

* addEventAndWait can also now trigger rules not defined in ECAL ([64e8369](https://devt.de///commit/64e83693a8fe114ac0414ad631e671edc84876cd))

### [1.4.5](https://devt.de///compare/v1.4.4...v1.4.5) (2021-01-31)


### Bug Fixes

* Allow adding fixed custom rules for CLI interpreter ([ea62e5a](https://devt.de///commit/ea62e5aaa92bf53ed4e95152f77f2ee7e699fac1))

### [1.4.4](https://devt.de///compare/v1.4.3...v1.4.4) (2021-01-15)


### Bug Fixes

* Better output for runtime errors ([6d8aea2](https://devt.de///commit/6d8aea29084102d267eb7e132b5d83d69c27a356))
* Better type conversion of nested structures and a new type function to inspect objects ([585ba90](https://devt.de///commit/585ba9018e1c3a36d237f1f40e3602240149c41f))
* First character of generated stdlib functions should be lower case. ([0f25d80](https://devt.de///commit/0f25d80738dbef6a681c0a6a36aba3e3fbf4c8d3))
* Including NaN and Inf functions in stdlib by default ([6f56203](https://devt.de///commit/6f56203e580ff8f22678748721efdbcb2a5549d9))
* More math functions and constants ([cc92319](https://devt.de///commit/cc923195bd67a5da93b1c7576c25d1b5ff2d0cba))

### [1.4.3](https://devt.de///compare/v1.4.2...v1.4.3) (2021-01-14)


### Bug Fixes

* Adding timestamp functions ([7ca66c5](https://devt.de///commit/7ca66c528a71f7a315c7d4cb914c2386f9adf5f5))

### [1.4.2](https://devt.de///compare/v1.4.1...v1.4.2) (2021-01-10)


### Bug Fixes

* except can now raise new errors / signals ([1550304](https://devt.de///commit/1550304c4c53916badbb005c400ac58517a82fc9))
* Try supports now otherwise blocks ([4a29d46](https://devt.de///commit/4a29d46f3ab4ef655131194f0a60e3e8fb62fafa))

### [1.4.1](https://devt.de///compare/v1.4.0...v1.4.1) (2021-01-10)


### Bug Fixes

* Pretty printer preserves some newlines to structure code. ([3320074](https://devt.de///commit/33200745a8331249e6edc05ef93a94fed398c7ee))

## [1.4.0](https://devt.de///compare/v1.3.3...v1.4.0) (2021-01-09)


### Features

* Better pretty printer ([18e42f7](https://devt.de///commit/18e42f771dc386a8a6cfb9329f59c5a246934be0))

### [1.3.3](https://devt.de///compare/v1.3.2...v1.3.3) (2021-01-01)


### Bug Fixes

* Not stopping debug server if not running in interactive mode ([4d62c35](https://devt.de///commit/4d62c353384d5cc718b891540abf0e5e344d9326))

### [1.3.2](https://devt.de///compare/v1.3.1...v1.3.2) (2020-12-30)


### Bug Fixes

* Adding error display when reloading interpreter state ([f5cc392](https://devt.de///commit/f5cc392a9cd4dd2a1d743fd9f1bc54bb71f1484a))

### [1.3.1](https://devt.de///compare/v1.3.0...v1.3.1) (2020-12-30)


### Bug Fixes

* Stopping and starting the processor when loading the initial file. ([981956f](https://devt.de///commit/981956ff9309b8395081d25a2c5711d872818015))

## [1.3.0](https://devt.de///compare/v1.2.0...v1.3.0) (2020-12-29)


### Features

* Adding conversion helper for JSON objects ([1050423](https://devt.de///commit/1050423c453169f22da029a52f131e2d3054a1f1))

## [1.2.0](https://devt.de///compare/v1.1.0...v1.2.0) (2020-12-26)


### Features

* Adding conversion helper for JSON objects ([0454de3](https://devt.de///commit/0454de30f32b70cb936675c86bf6bdafd4e4bc3b))

## [1.1.0](https://devt.de///compare/v1.0.4...v1.1.0) (2020-12-13)


### Features

* Adding plugin support to ECAL ([56be402](https://devt.de///commit/56be402e464b5f9574295e717a4cebc382852c26))


### Bug Fixes

* Pack can now run under Windows / adjusted example scripts ([6f24339](https://devt.de///commit/6f243399ea5c3042ad1092b94a6372e0fd55a5e0))

### [1.0.4](https://devt.de///compare/v1.0.3...v1.0.4) (2020-12-07)

### [1.0.3](https://devt.de///compare/v1.0.2...v1.0.3) (2020-12-07)


### Bug Fixes

* Test fixes ([9075df4](https://devt.de///commit/9075df4c630a25db597a3c48c74d230383dfd5e8))
