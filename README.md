# refbuilder
CHM-like index and contents generator for directories of HTML files. Written by Stiletto and open-sourced on terms of MIT license.

Uses [JSTree](https://github.com/vakata/jstree), [jQuery](http://jquery.com) and [icons by Benjamin STAWARZ](https://www.iconfinder.com/butterflytronics).

## Installing

```sh
go get github.com/stiletto/refbuilder
```
## Using
```sh
refbuilder /path/to/directory
```
## Building (if you intend to modify refbuilder)
If you want to JS files to be minified - download Google Closure Compiler and put in `compiler.jar` in this directory.
```sh
bash makeassets.sh # downloads javascript dependencies and makes assets directory.
bash makestatik.sh # builds statik package for embedding in resulting binary
cd cmd/refbuilder && go build
```
