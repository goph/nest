# Change Log


All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased]


## [0.5.1] - 2018-01-17

### Fixed

- Empty string flags and environment variables now fall back to the zero value of the type to avoid parsing issues


## [0.5.0] - 2018-01-13

### Added

- `SetOutput` method for setting an output writer on the Configurator

### Changed

- Improved usage output: added environment variables


## [0.4.0] - 2018-01-02

### Added

- `isExported` util function to avoid importing `ast` package
- Tests for embedded structs
- Support for types implementing `encoding.TextUnmarshaler`
- Support for types implementing `Decoder`


## [0.3.0] - 2017-12-21

### Added

- Missing global `SetName` function
- `SetArgs` to set command line arguments manually
- Add struct tag constant values
- Support child structs
- `prefix` tag to configure custom prefix for child structs

### Changes

- Refactor internals


## [0.2.0] - 2017-12-19

### Added

- `time.Duration` support
- `split_words` tag support for splitting camel cased field names to environment variable and flag names

### Changed

- `viper.Viper` instance is maintained during the configurator lifecycle and is not recreated
- Use first argument as application name when not provided
- Set application name globally when not provided

### Fixed

- Panic when help is requested ([#1](https://github.com/goph/nest/issues/1))
- Application name when help is displayed ([#2](https://github.com/goph/nest/issues/2))


## 0.1.0 - 2017-12-18

- Initial release


[Unreleased]: https://github.com/goph/nest/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/goph/nest/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/goph/nest/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/goph/nest/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/goph/nest/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/goph/nest/compare/v0.1.0...v0.2.0
