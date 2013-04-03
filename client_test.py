import riak

# pb port is actually not used
client = riak.RiakClient(nodes=[{"host": "127.0.0.1", "http_port": "8198", "pb_port": "8197"}])

bucket = client.bucket("test")
obj = bucket.new("test", "yay")
