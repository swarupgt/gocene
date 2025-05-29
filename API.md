# API Documentation

Below are the currently supported APIs

###  1. Create Index
POST `/create_index`
```JSON
{
    "name": "index_name"
}
```

### 2. Add Document
POST `/<index_name>/add_document`
```JSON
{
    "data": {
        "field1": "text1",
        "field2": "some text and some more text about swords and steel and other cool stuff"
    }
}
```

### 3. Search (Full Text)
POST `/<index_name>/search`
```JSON
{
    "search_field": "field_name",
    "search_phrase": "some words to search for"
}
```

### 4. Get Document
POST `/<index_name>/get_document`
```JSON
{
   "doc_id": 1
}
```