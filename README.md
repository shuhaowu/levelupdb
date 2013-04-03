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

To be written

Usage and Configurations
------------------------

To be written

Technical Details
-----------------

To be written

API Documentations
------------------

To be written. Consult Riak
