# Change Log


All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

- `time.Duration` support

### Changed

- `viper.Viper` instance is maintained during the configurator lifecycle and is not recreated
- Use first argument as application name when not provided
- Set application name globally when not provided

### Fixed

- Panic when help is requested (#1)
- Application name when help is displayed (#2)


## 0.1.0 - 2017-12-18

- Initial release


[Unreleased]: https://github.com/goph/fxt/compare/v0.1.0...HEAD
