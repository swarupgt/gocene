# gocene
A rudimentary implementation of an indexing and search engine in Go, kinda inspired by Apache Lucene.

## TODO
- binary search for getDoc and modifyDoc
- load persistent indices everytime server starts (currently does not, support exists tho)
- test the Full text search functionality (done)
- add logging 
- API for search and addition of documents (partially done)
- Docker image

## Notes
- Consider dirty reads when index is currently being dropped