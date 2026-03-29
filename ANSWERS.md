# Answers

## Q1
It will be a new mutex in each function call. So it wouldn't block operations on map. 

## Q2
Goroutine 1 lock A and wait on B while goroutine 2 locks B and waits on A. This is a deadlock. Global lock for multi‑item operations would be appropriate.

## Q3
Another goroutine can change stock in between, so you can oversell or go negative.

## Q4
No. The race detector can miss races. We can face race in production.
