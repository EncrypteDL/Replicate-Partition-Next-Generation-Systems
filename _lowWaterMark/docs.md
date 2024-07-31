# Low Water Mark 

In distributed systems, a Low-Water Mark (LWM) refers to a specific point within a write-ahead log that indicates the safe minimum threshold for discarding older log entries.  Think of it as a marker in the sand.
Overall. The Low-Water Mark (LWM) in the context of distributed systems and databases is a threshold that indicates the minimum level or point below which certain actions are triggered. Specifically, it is used to manage resources, ensure consistency, and maintain system performance by providing a reference for cleanup operations and checkpoints.

## Here's how it works:
- Distributed systems often rely on write-ahead logs to ensure data consistency and recover from failures. These logs record all the changes made to the system.
- However, storing all the logs forever can be impractical due to storage limitations.
- The Low-Water Mark helps determine which parts of the log can be safely discarded without compromising data integrity.
- Resource Management: LWM helps in managing resources like memory, storage, and processing power by indicating when cleanup or compaction should occur.
- Consistency: In distributed databases, LWM can indicate the point up to which all nodes agree on the state of data, ensuring consistency.
- Checkpointing: It serves as a reference point for creating checkpoints, which are snapshots of the system state that can be used for recovery purposes.
- Garbage Collection: In systems with log-structured storage, LWM is used to trigger garbage collection processes to reclaim space.

### There are two main uses of Low-Water Marks:
1. **Log Management:** The LWM acts as a signal to the logging system. Any entries before the LWM can be safely deleted because they've already been processed and the system has reached a stable state beyond that point.
2. **Data Synchronization:** In scenarios where multiple systems need to stay in sync, the LWM can be used to identify changes that have occurred since the last synchronization point. This allows systems to efficiently retrieve only the most recent updates, improving efficiency.

## Problems Addressed by Low-Water Mark
- Resource Overuse: Prevents excessive resource usage by triggering cleanup operations when necessary.
- Inconsistency: Helps maintain consistency by ensuring all nodes or replicas are synchronized up to a certain point.
- Performance Degradation: Prevents performance degradation by managing resource thresholds and triggering appropriate actions.

## Solution Provided by Low-Water Mark
- Efficient Resource Management: By defining a lower threshold, the system can efficiently manage and reclaim resources.
- Consistency Maintenance: Ensures that the system state is consistent across distributed nodes or replicas.
- System Recovery: Provides a reference point for creating checkpoints, aiding in quick recovery from failures.

-------------------
## Use Cases of Low-Water Mark

| Use Case                         | Description                                                                        |
|----------------------------------|------------------------------------------------------------------------------------|
| **Distributed Databases**        | Ensures consistency and aids in garbage collection of outdated log entries.         |
| **Memory Management**            | Used in operating systems and databases to manage buffer pools and memory allocation. |
| **Data Replication**             | Maintains consistent state across replicas in distributed systems.                 |
| **Log-Structured File Systems**  | Triggers garbage collection to reclaim space from outdated log entries.            |

### getting started
```
func main() {
    logManager := NewLogManager()
    
    logManager.AddEntry(LogEntry{Index: 1, Data: "Entry 1"})
    logManager.AddEntry(LogEntry{Index: 2, Data: "Entry 2"})
    logManager.AddEntry(LogEntry{Index: 3, Data: "Entry 3"})

    fmt.Println("Entries before cleanup:", logManager.entries)

    logManager.SetLowWaterMark(2)

    fmt.Println("Entries after cleanup:", logManager.entries)
}
```