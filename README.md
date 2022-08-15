# consistent-hashing-service-provider

**consistent-hashing-service-provider** is designed to provide consistent-hashing service in distributed manner supported by replication + oplog.

![](doc/arch.jpeg) 

## Get Started

### Prerequisites

```text
1. debian/ubuntu linux/x86-64 release
2. go1.18+ linux/amd64 or higher
```

### Installation

#### Clone

* Clone this repo to your local machine using https://github.com/amazingchow/consistent-hashing-service-provider.git.

#### Setup

```shell
# build the binary
make build

# start one master node
./consistent-hashing-service-provider --id="localhost:18081" --conf=conf/master.json --verbose=true

# start two slave nodes
./consistent-hashing-service-provider --id="localhost:18082" --conf=conf/slave01.json --verbose=true
./consistent-hashing-service-provider --id="localhost:18083" --conf=conf/slave02.json --verbose=true

# use grpcurl
grpcurl -plaintext -d '{"node": {"uuid": "192.168.1.125"}}' localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Add
grpcurl -plaintext -d '{"node": {"uuid": "192.168.1.126"}}' localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Add
grpcurl -plaintext -d '{"node": {"uuid": "192.168.1.127"}}' localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Add
grpcurl -plaintext -d '{"node": {"uuid": "192.168.1.128"}}' localhost:18082 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Add
grpcurl -plaintext -d '{"node": {"uuid": "192.168.1.129"}}' localhost:18083 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Add
grpcurl -plaintext localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext localhost:18082 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext localhost:18083 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext -d '{"uuid": "192.168.1.126"}' localhost:18082 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Delete
grpcurl -plaintext -d '{"uuid": "192.168.1.128"}' localhost:18083 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/Delete
grpcurl -plaintext localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext localhost:18082 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext localhost:18083 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/List
grpcurl -plaintext -d '{"key": {"name": "foo"}}' localhost:18082 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/MapKey
grpcurl -plaintext -d '{"key": {"name": "bar"}}' localhost:18083 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/MapKey
grpcurl -plaintext -d '{"key": {"name": "summychou"}}' localhost:18081 amazingchow.photon_dance_consistent_hashing.ConsistentHashingService/MapKey
```

## Reference

* [Consistent Hashing and Random Trees: Distributed Caching Protocols for Relieving Hot Spots on the World Wide Web](https://www.akamai.com/us/en/multimedia/documents/technical-publication/consistent-hashing-and-random-trees-distributed-caching-protocols-for-relieving-hot-spots-on-the-world-wide-web-technical-publication.pdf)
* [Consistent Hashing with Bounded Loads](https://arxiv.org/pdf/1608.01350.pdf)
* [A Fast, Minimal Memory, Consistent Hash Algorithm](https://arxiv.org/pdf/1406.2294.pdf)

## Contributing

### Step 1

* 🍴 Fork this repo!

### Step 2

* 🔨 HACK AWAY!

### Step 3

* 🔃 Create a new PR using https://github.com/amazingchow/consistent-hashing-service-provider/compare!

## Support

* Reach out to me at <jianzhou42@163.com>.

## License

* This project is licensed under the MIT License - see the **[MIT license](http://opensource.org/licenses/mit-license.php)** for details.
