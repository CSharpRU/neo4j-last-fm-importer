## Synopsis

This is a simple application for loading top tags and tracks into Neo4j from Last.fm.

This app also does simple syntactic analysis for track description to get it emotional context.

## Motivation

It is fun :) 

I love Neo4j, graphs and their visualization so much :)

## Installation

1. Get source code and depencencies:
```
git clone git@github.com:CSharpRU/neo4j-last-fm-importer.git
go get gopkg.in/jmcvetta/neoism.v1
go get github.com/imdario/mergo
go get github.com/CSharpRU/lastfm-go/lastfm
```
2. Change configuration in `config.yml` file.
3. Execute `go run main.go` to run application

## Tests

TBD.

## Contributors

Feel free to expand and extend this source code. Also you can contact with me via GitHub or [Twitter](https://twitter.com/CSharpGL) :)

## License

See [LICENSE](https://github.com/CSharpRU/neo4j-last-fm-importer/blob/master/LICENSE).
