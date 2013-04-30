Levelupdb
=========

A single node high performance key-value database server based on leveldb.
Written with Go.

Levelupdb is API compatible with the Riak HTTP API (PBC is planned) and any
existing Riak libraries should be in theory compatible with Levelupdb.

Use case: If you have a low end box and want to host some application that one
day you might want to scale and the problem is compatible with a riak like
solution.

License: GPLv3

    Latest performance test (512MB, no swap, Vagrant Box, Debian Wheezy, Intel(R) Core(TM) i5-2410M CPU @ 2.30GHz, 1 core, SanDisk Extreme SSD 240GB)
    Tests ran independently of each other. The system was reset between HTTP, PBC, and Levelup tests.
    Configurations are all on their defaults.

    riak (1.3.1 2013-04-03) Debian x86_64
    levelupdb 0.1 (master), go1.0.3 Debian x86_64

    Benchmarking Riak HTTP Insert
    100% complete...
    Total time: 90.14 seconds

    Benchmarking Riak HTTP Fetch
    100% complete...
    Total time: 59.98 seconds

    Benchmarking Riak HTTP Delete
    100% complete...
    Total time: 143.51 seconds

    Benchmarking Riak PBC InserBenchmarking Riak PBC Insert
    100% complete...
    Total time: 48.14 seconds

    Benchmarking Riak PBC Fetch
    100% complete...
    Total time: 31.27 seconds

    Benchmarking Riak PBC Delete
    100% complete...
    Total time: 96.25 seconds

    Benchmarking Levelupdb Insert
    100% complete...
    Total time: 19.77 seconds

    Benchmarking Levelupdb Fetch
    100% complete...
    Total time: 14.71 seconds

    Benchmarking Levelupdb Delete
    100% complete...
    Total time: 11.97 seconds

    Document Size: 8000 bytes
    Number of Documents: 10000

    Typical CPU Usage by Benchmarking Script (Riak): 30%
    Typical CPU Usage by Benchmarking Script (Levelupdb): 70%

Compatibility with Riak
-----------------------

Levelupdb is designed to be API compatible with Riak, which means that Levelupdb
already have support from all major languages (except for Go, which lacks an
HTTP Riak client). There are extra features such as write batches that only
exists within levelupdb (as of Riak 1.3). For now, Levelupdb only supports the
HTTP interface (**new riak format only**)

Remember, this is not a competition. This is a db that solves its own areas and
allow you to easily transition to Riak :P

Another thing is if any of these differences causes a failure of a client, it
must be fixed.

There are differences between Riak and Levelupdb:

 1. **Consistency is guarenteed**: since Levelupdb runs on a single node,
    consistency is guarenteed. A 200 is only returned after the write has
    completed. See point below for conflict resolution
 2. **Conflict resolution is different**: At this time the conflict resolution
    is last write wins. This means there is no siblings or anything like that.
     I hope to add some sort of vector clock system in the future.
 3. **Different headers**: Some *non-essential* HTTP headers may be different.
    Such as `Server`. Some headers may be nonexistent in levelupdb, such as
    `Content-Length`.
 4. **Bucket properties are different/not available**: Certain bucket properties
    that's for distributed-ness are not available.
 5. **SOLR Search is not available**: Maybe down the line..
 6. **Map reduce is not yet available**: This is a planned feature. Erlang map
    reduce probably will never be available. If you rely on this feature, it
    might not be a good idea to use this in place of riak (riak can run on
    lowendboxes as well)
 7. **Designed to run on a single node**: Riak is designed to run on a cluster.
    It's performance on a single node may not be optimal. Levelupdb is designed
    to run on lowendboxes and small VPSes. It makes hosting your side projects
    painless. **It wants to host your side projects**.
 8. **List keys and buckets is not a big deal**: Listing keys and buckets do not
    cost as much as they do in Riak. Listing buckets just takes a directory list
    and listing all keys only iterate through that specific bucket.

Some more technical differences:

 - Delete request will not return 404 if the content is not found, but
   204 instead.
 - Since we don't deal with conflicts, the vector clock for each object is
   always the same (for all objects.)

Rational
--------

This is essentially my brain storming section.

Levelupdb is designed as a high performance database server. The main
goal is to have a database server that has a minimal footprint but yet can still
perform well. The target audience is low end boxes and side projects.

A secondary goal, but very important, is the compatibility with Riak. This means
that applications developed for levelupdb should be compatible with Riak with
minimum porting effort (preferably levelupdb must be able to use riak clients).

This means a couple of things:

 1. **Conflict is minimal**: The scope of levelupdb is low traffic side
    side projects running on LEB (students anyone? :D). This means that there
    should not be a lot of conflicts and if there are conflicts, last write
    win is acceptable.
 2. **Low footprint**: Should not use a lot of RAM.
 3. **Ease of use**: Should be easily hooked up with minimal amount of
    configuration and effort.

Installation
------------

Levelupdb is only tested on Linux. Should run where leveldb and snappy compiles.

To install levelupdb, you need to install snappy and leveldb. Use the latest
version and you have to manually move the library files for leveldb into your
system's dev library location.

`go build` should build the binary `levelupdb`

`go install` so you can have it in your `GOPATH`. To run it just use
`./levelupdb`. Daemonize using your os.

Usage and Configurations
------------------------

To be written

Technical Details
-----------------

To be written

API Documentations
------------------

To be written. Consult Riak
