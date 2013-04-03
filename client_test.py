import riak
import unittest

# This code here is Apache 2 because basho people wrote it :D
# I may or may not have contributed before?

class KVTests(unittest.TestCase):
  def setUp(self):
    self.client = riak.RiakClient(nodes=[{"host": "127.0.0.1", "http_port": "8198", "pb_port": "8197"}])
    self.bucket_name = "test"
    self.key_name = "test"

  def test_delete(self):
    bucket = self.client.bucket(self.bucket_name)
    obj = bucket.new(self.key_name, 123)
    obj.store()
    obj = bucket.get(self.key_name)
    self.assertTrue(obj.exists)

    obj.delete()
    obj.reload()
    self.assertFalse(obj.exists)

if __name__ == "__main__":
  unittest.main()