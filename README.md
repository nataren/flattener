Dependencies
============

1. [Go](https://golang.org/ "Go Programming Language")


How to run this
===============

1. Using a bash terminal issue the following command: ```go run press.go --dir={PATH_TO_DIRECTORY_WITH_LOG_FILES}```

Output
======

1. You will find newly created {FILENAME}.log.csv files based on the original log files

Todo
====

1. Add support for anything else that we are missing :-)

Notes
=====

1. The ```event/request/parameters/param``` elements are concatenated as a single value for the column ```request.parameters``` with the format ```paramName1:paramValue1;paramName2:paramValue2;...```
2. The ```event/tags-added/tag``` elements are concatenaed as a single value for the column ```tags.added``` with the format ```tagName1^tagValue1;tagName2^tagValue2;...```
3. The ```event/tags-removed/tag``` elements are concatenaed as a single value for the column ```tags.removed``` with the format ```tagName1^tagType1;tagName2^tagType2;...```

