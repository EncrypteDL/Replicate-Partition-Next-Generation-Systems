# Quorum DS
In distributed systems, a quorum is the minimum number of votes or acknowledgments required for a transaction to be considered valid or for a decision to be made. This concept ensures that operations are consistent and reliable, even in the presence of network partitions or node failures.

### Key Concepts

1. **Quorum-based Voting**: A method of ensuring consistency by requiring a majority of nodes to agree on an operation.
2. **Read Quorum**: The minimum number of nodes that must respond to a read request.
3. **Write Quorum**: The minimum number of nodes that must acknowledge a write request.
4. **Quorum Formula**: Generally, to achieve consistency, the sum of read and write quorums should be greater than the total number of nodes.

### Problems Addressed by Quorum

1. **Consistency**: Ensures that all nodes have a consistent view of the data.
2. **Fault Tolerance**: Allows the system to function correctly even when some nodes fail.
3. **Network Partitions**: Helps in making progress despite network partitions by relying on a subset of nodes.

### Solutions Provided by Quorum

- **Consistency**: By requiring a majority of nodes to agree, quorum ensures that all nodes have the same view of the data.
- **Availability**: Ensures the system remains available and operational, even with node failures.
- **Partition Tolerance**: Maintains consistency and availability during network partitions.

## Example & Use Cases of Quorum

All the consensus implementations like Zab, Raft, Paxos are quorum based.

• Even in systems which don’t use consensus, quorum is used to make sure the latest update is available to at least one server in case of failures or network partition. For instance, in databases like Cassandra, a database update can be configured to return success only after a majority of the servers have updated the record successfully.

### Use Cases

1. **Distributed Databases**: Ensures consistency and reliability in distributed database operations.
2. **Consensus Algorithms**: Used in consensus protocols like Paxos and Raft to achieve agreement among distributed nodes.
3. **Distributed Locking**: Ensures that locks are acquired and released consistently across a distributed system.
4. **Blockchain**: Ensures that transactions are validated by a majority of nodes in the network.