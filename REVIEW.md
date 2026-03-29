# Race Condition Review

## Race Condition 1: GetStock (read) vs Reserve/ReserveMultiple (write)
- Code: `GetStock` reads `product.Stock` while `Reserve`/`ReserveMultiple` write it.
- What happens: concurrent read/write of the same field without synchronization.
- Production scenario: goroutine A calls `GetStock("p1")` while goroutine B calls `Reserve("p1", 1)`; the read races with the write.
- Fix approach: protect `Stock` with a shared `sync.RWMutex` (`RLock` for reads, `Lock` for writes).

## Race Condition 2: Reserve check-and-update is not atomic
- Code: `if product.Stock < quantity { ... }` then `product.Stock -= quantity` without a lock.
- What happens: two goroutines can both pass the check and both decrement, overselling stock.
- Production scenario: stock=1, goroutine A and B both call `Reserve("p1", 1)` simultaneously; both pass check, both decrement, resulting stock=-1.
- Fix approach: wrap the check and update in the same exclusive critical section.

## Race Condition 3: ReserveMultiple check phase separated from update phase
- Code: first loop checks all items, second loop updates them, with no lock.
- What happens: stock can change between the loops, so the update can oversell or partially apply based on stale checks.
- Production scenario: goroutine A checks items (all OK), goroutine B reserves one item, then goroutine A proceeds to subtract and oversells.
- Fix approach: lock once for the entire operation and perform check + update atomically.

## Race Condition 4: SafeReserve uses a per-call mutex
- Code: `var mu sync.Mutex` declared inside `SafeReserve`.
- What happens: each call uses a different mutex, so goroutines do not synchronize with each other at all; the lock has no effect.
- Production scenario: goroutine A and B both call `SafeReserve`; they lock different mutex instances and still race on `product.Stock`.
- Fix approach: use a shared mutex (field on the service) to synchronize all goroutines.
