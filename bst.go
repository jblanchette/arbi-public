package main

import (
    "fmt"
)

type OrderTree struct {
    Left  *OrderTree
    Value BookEntry
    Right *OrderTree
}

// Walk traverses a OrderTree depth-first,
// sending each Value on a channel.
func WalkOrderTree(t *OrderTree, ch chan BookEntry) {
    if t == nil {
        return
    }
    WalkOrderTree(t.Left, ch)
    ch <- t.Value
    WalkOrderTree(t.Right, ch)
}

// Walker launches Walk in a new goroutine,
// and returns a read-only channel of values.
func OrderTreeWalker(t *OrderTree) <-chan BookEntry {
    ch := make(chan BookEntry)
    go func() {
        WalkOrderTree(t, ch)
        close(ch)
    }()
    return ch
}


var equalityEpsilon float64 = 0.00000001
func floatWithin(a, b, epsilon float64) bool {
    if ((a - b) < epsilon && (b - a) < epsilon) {
        return true
    }
    return false
}

func FindClosestLeaf(t *OrderTree, rate float64, tolerance float64) (BookEntry, bool) {
    c := OrderTreeWalker(t)
    var lowVal, lowStart float64 = 1000000.0, 1000000.0
    var lowBook BookEntry

    for {
        v, ok := <-c

        if floatWithin(v.rate, rate, equalityEpsilon) {
            fmt.Println("Found 'exact match'")
            return v, true
        }

        if floatWithin(v.rate, rate, tolerance) {
            if v.rate < lowVal {
                lowVal = v.rate
                lowBook = v
            }
        }

        if !ok {
            break
        }
    }

    if lowVal != lowStart {
        fmt.Println("Found 'approximate match'")
        return lowBook, true
    }

    fmt.Println("Did not find a match.")
    return BookEntry{}, false
}

func InsertOrderTreeNode(t *OrderTree, book BookEntry) *OrderTree {
    if t == nil {
        return &OrderTree{nil, book, nil}
    }
    if book.rate < t.Value.rate {
        t.Left = InsertOrderTreeNode(t.Left, book)
        return t
    }
    t.Right = InsertOrderTreeNode(t.Right, book)
    return t
}
