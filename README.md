# gocene
A distributed search engine in Go, inspired by Apache Lucene (kinda). I plan to integrate it with Kubernetes.

## Design

The service runs as a Raft cluster, with a single writer.

Each Store object contains a list of indices that are searchable. Each Index object has a list of immutable segments and an active segment. New documents added are only to the active segment, and this is flushed to the list of immutable segments once it reaches a certain document count. This is to ensure concurrency when searching on an index.

Gocene's indices store the documents themselves in Minio, but any other S3 compatible buckets can also be used. 

You can currently - 
1. create index
2. add documents
3. search full text
<!--4. get a document-->

Search only supports a single field for now :/ It's a simple engine, using only the frequency of words as the score (lol). 

API Documentation -> [HERE](./API.md)

## TODO
- explicit node joining as part of raft logs
- remove useless passthroughs like the API service/controller 
- td-idf for ranking
- tests
