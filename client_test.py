# -*- coding: utf-8 -*-
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

  def test_store_and_get(self):
    bucket = self.client.bucket(self.bucket_name)
    rand = 42
    obj = bucket.new('foo', rand)
    obj.store()
    obj = bucket.get('foo')
    self.assertTrue(obj.exists)
    self.assertEqual(obj.bucket.name, self.bucket_name)
    self.assertEqual(obj.key, 'foo')
    self.assertEqual(obj.data, rand)

    # unicode objects are fine, as long as they don't
    # contain any non-ASCII chars
    self.client.bucket(unicode(self.bucket_name))
    self.assertRaises(TypeError, self.client.bucket, u'búcket')
    self.assertRaises(TypeError, self.client.bucket, 'búcket')

    bucket.get(u'foo')
    self.assertRaises(TypeError, bucket.get, u'føø')
    self.assertRaises(TypeError, bucket.get, 'føø')

    self.assertRaises(TypeError, bucket.new, u'foo', 'éå')
    self.assertRaises(TypeError, bucket.new, u'foo', 'éå')
    self.assertRaises(TypeError, bucket.new, 'foo', u'éå')
    self.assertRaises(TypeError, bucket.new, 'foo', u'éå')

  def test_blank_binary_204(self):
    bucket = self.client.bucket(self.bucket_name)

    # this should *not* raise an error
    obj = bucket.new('foo2', encoded_data='', content_type='text/plain')
    obj.store()
    obj = bucket.get('foo2')
    self.assertTrue(obj.exists)
    self.assertEqual(obj.encoded_data, '')

  def test_secondary_index_store(self):
    # Create a new object with indexes...
    bucket = self.client.bucket(self.bucket_name)
    rand = 13337
    obj = bucket.new('mykey1', rand)
    obj.add_index('field1_bin', 'val1a')
    obj.add_index('field1_int', 1011)
    obj.store()

    # Retrieve the object, check that the correct indexes exist...
    obj = bucket.get('mykey1')
    self.assertEqual(['val1a'], [y for (x, y) in obj.indexes
                                 if x == 'field1_bin'])
    self.assertEqual([1011], [y for (x, y) in obj.indexes
                              if x == 'field1_int'])

    # Add more indexes and save...
    obj.add_index('field1_bin', 'val1b')
    obj.add_index('field1_int', 1012)
    obj.store()

    # Retrieve the object, check that the correct indexes exist...
    obj = bucket.get('mykey1')
    self.assertEqual(['val1a', 'val1b'],
                     sorted([y for (x, y) in obj.indexes
                             if x == 'field1_bin']))
    self.assertEqual([1011, 1012],
                     sorted([y for (x, y) in obj.indexes
                             if x == 'field1_int']))

    self.assertEqual(
        [('field1_bin', 'val1a'),
         ('field1_bin', 'val1b'),
         ('field1_int', 1011),
         ('field1_int', 1012)
         ], sorted(obj.indexes))

    # Delete an index...
    obj.remove_index('field1_bin', 'val1a')
    obj.remove_index('field1_int', 1011)
    obj.store()

    # Retrieve the object, check that the correct indexes exist...
    obj = bucket.get('mykey1')
    self.assertEqual(['val1b'], sorted([y for (x, y) in obj.indexes
                                        if x == 'field1_bin']))
    self.assertEqual([1012], sorted([y for (x, y) in obj.indexes
                                     if x == 'field1_int']))

    # Check duplicate entries...
    obj.add_index('field1_bin', 'val1a')
    obj.add_index('field1_bin', 'val1a')
    obj.add_index('field1_bin', 'val1a')
    obj.add_index('field1_int', 1011)
    obj.add_index('field1_int', 1011)
    obj.add_index('field1_int', 1011)

    self.assertEqual(
        [('field1_bin', 'val1a'),
         ('field1_bin', 'val1b'),
         ('field1_int', 1011),
         ('field1_int', 1012)
         ], sorted(obj.indexes))

    obj.store()
    obj = bucket.get('mykey1')

    self.assertEqual(
        [('field1_bin', 'val1a'),
         ('field1_bin', 'val1b'),
         ('field1_int', 1011),
         ('field1_int', 1012)
         ], sorted(obj.indexes))

    # Clean up...
    bucket.get('mykey1').delete()


  def test_secondary_index_query(self):
    bucket = self.client.bucket(self.bucket_name)

    bucket.\
        new('mykey1', 'data1').\
        add_index('field1_bin', 'val1').\
        add_index('field2_int', 1001).\
        store()
    bucket.\
        new('mykey2', 'data1').\
        add_index('field1_bin', 'val2').\
        add_index('field2_int', 1002).\
        store()
    bucket.\
        new('mykey3', 'data1').\
        add_index('field1_bin', 'val3').\
        add_index('field2_int', 1003).\
        store()
    bucket.\
        new('mykey4', 'data1').\
        add_index('field1_bin', 'val4').\
        add_index('field2_int', 1004).\
        store()

    # Test an equality query...
    results = bucket.get_index('field1_bin', 'val2')
    self.assertEquals(1, len(results))
    self.assertEquals('mykey2', str(results[0]))

    # Test a range query...
    results = bucket.get_index('field1_bin', 'val2', 'val4')
    vals = set([str(key) for key in results])
    self.assertEquals(3, len(results))
    self.assertEquals(set(['mykey2', 'mykey3', 'mykey4']), vals)

    # Test an equality query...
    results = bucket.get_index('field2_int', 1002)
    self.assertEquals(1, len(results))
    self.assertEquals('mykey2', str(results[0]))

    # Test a range query...
    results = bucket.get_index('field2_int', 1002, 1004)
    vals = set([str(key) for key in results])
    self.assertEquals(3, len(results))
    self.assertEquals(set(['mykey2', 'mykey3', 'mykey4']), vals)

    # Clean up...
    bucket.get('mykey1').delete()
    bucket.get('mykey2').delete()
    bucket.get('mykey3').delete()
    bucket.get('mykey4').delete()

  def test_secondary_index_invalid_name(self):
    bucket = self.client.bucket(self.bucket_name)

    with self.assertRaises(riak.RiakError):
      bucket.new('k', 'a').add_index('field1', 'value1')

  def test_store_and_get_links(self):
    # Create the object...
    bucket = self.client.bucket(self.bucket_name)
    bucket.new(key="test_store_and_get_links", encoded_data='2',
               content_type='application/octet-stream') \
        .add_link(bucket.new("foo1")) \
        .add_link(bucket.new("foo2"), "tag") \
        .add_link(bucket.new("foo3"), "tag2!@#%^&*)") \
        .store()
    obj = bucket.get("test_store_and_get_links")
    links = obj.links
    self.assertEqual(len(links), 3)
    for bucket, key, tag in links:
        if (key == "foo1"):
            self.assertEqual(bucket, self.bucket_name)
        elif (key == "foo2"):
            self.assertEqual(tag, "tag")
        elif (key == "foo3"):
            self.assertEqual(tag, "tag2!@#%^&*)")
        else:
            self.assertEqual(key, "unknown key")

  def test_set_links(self):
    # Create the object
    bucket = self.client.bucket(self.bucket_name)
    o = bucket.new(self.key_name, 2)
    o.links = [(self.bucket_name, "foo1", None),
               (self.bucket_name, "foo2", "tag"),
               ("bucket", "foo2", "tag2")]
    o.store()
    obj = bucket.get(self.key_name)
    links = sorted(obj.links, key=lambda x: x[1])
    self.assertEqual(len(links), 3)
    self.assertEqual(links[0][1], "foo1")
    self.assertEqual(links[1][1], "foo2")
    self.assertEqual(links[1][2], "tag")
    self.assertEqual(links[2][1], "foo2")
    self.assertEqual(links[2][2], "tag2")

  def test_link_walking(self):
    pass
    # python client lacks http link walking

if __name__ == "__main__":
  unittest.main()