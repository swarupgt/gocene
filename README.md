# gocene
A rudimentary implementation of an indexing and search engine in Go, kinda inspired by Apache Lucene.

You can currently - 
1. create index
2. add documents
3. search full text
4. get a document

Search only supports a single field for now :/

API Documentation -> [HERE](./API.md)

## TODO
- segments and persistent indices
- tests 
- partitioning and replication for durability

## Notes
- Consider dirty reads when index is currently being dropped
- search only returns doc ids, make code cleaner for that