from __future__ import division

import riak
import random
import string
import time
from sys import stdout

DOCUMENT_SIZE = 8000
NUM_OF_DOCUMENTS = 10000

def generate_document(size):
  doc = ""
  for i in xrange(size):
    doc += random.choice(string.letters)
  return doc

def benchmark(func, num):
  one = num // 100
  percent = 0
  total = 0
  for i in xrange(num):
    start = time.time()
    func(i)
    total += time.time() - start

    if i % one == 0:
      percent += 1
      stdout.write("\r%d%% complete..." % percent)
      stdout.flush()

  stdout.write("\n")
  return total

def main():
  riakHttpClient = riak.RiakClient()
  riakPbcClient = riak.RiakClient(protocol="pbc")
  levelupClient = riak.RiakClient(nodes=[{"host": "127.0.0.1", "http_port": "8198", "pb_port": "8197"}])

  def display_result(time_taken, num):
    print "Total time: {0} seconds".format(round(time_taken, 2))

  document = generate_document(DOCUMENT_SIZE)

  def do_insert(bucket, i):
    obj = bucket.new("test" + str(i), encoded_data=document)
    obj.store()

  def do_get(bucket, i,):
    bucket.get("test"+str(i))

  def do_delete(bucket, i):
    bucket.new("test" + str(i)).delete()

  def create_func(c, f):
    bucket = c.bucket("test_bucket")
    def fn(i):
      f(bucket, i)
    return fn

  def benchmark_one(label, f):
    print "Benchmarking {0}".format(label)
    display_result(benchmark(f, NUM_OF_DOCUMENTS), NUM_OF_DOCUMENTS)
    print

  riakHttpCache = []
  riakPbcCache = []
  levelupCache = []

  #benchmark_one("Riak HTTP Insert", create_func(riakHttpClient, do_insert))
  #benchmark_one("Riak HTTP Fetch", create_func(riakHttpClient, do_get))
  #benchmark_one("Riak HTTP Delete", create_func(riakHttpClient, do_delete))
  riakHttpCache = []

  #benchmark_one("Riak PBC Insert", create_func(riakPbcClient, do_insert))
  #benchmark_one("Riak PBC Fetch", create_func(riakPbcClient, do_get))
  #benchmark_one("Riak PBC Delete", create_func(riakPbcClient, do_delete))
  riakPbcCache = []

  benchmark_one("Levelupdb Insert", create_func(levelupClient, do_insert))
  benchmark_one("Levelupdb Fetch", create_func(levelupClient, do_get))
  benchmark_one("Levelupdb Delete", create_func(levelupClient, do_delete))
  levelupCache = []

  print "Document Size: {0} bytes".format(DOCUMENT_SIZE)
  print "Number of Documents: {0}".format(NUM_OF_DOCUMENTS)

if __name__ == "__main__":
  main()
