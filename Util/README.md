# Util

## util functions
// todo


## DS
set:

ref: [golang-set](https://github.com/deckarep/golang-set)
```go
sub := Util.NewSet[int]()
sub.Add(1)
sub.Add(2)
sub.Add(3)

sup := sub.Clone()

fmt.Println("s is subset of sup:", sub.IsSubOf(sup))
fmt.Println("s is proper subset of sup:", sub.IsProperSubOf(sup))
sup.Add(4)
fmt.Println("s is subset of sup:", sub.IsSubOf(sup))
fmt.Println("s is proper subset of sup:", sub.IsProperSubOf(sup))

fmt.Println("to slice:", sub.ToSlice())
fmt.Println("items:", sub.Items())

fmt.Println("sup contains sub:", sup.ContainsAll(sub.Items()...))
fmt.Println("sub contains sup:", sub.ContainsAll(sup.Items()...))

fmt.Println("diff:", sup.Diff(sub))

// Output:
s is subset of sup: true
s is proper subset of sup: false
s is subset of sup: true
s is proper subset of sup: true
to slice: [1 2 3]
items: [1 2 3]
sup contains sub: true
sub contains sup: false
diff: set[4]
```

pair:

```go
p := Util.Pair[int, string]{}
p.Set(1, "a")
if p.First != 1 || p.Second != "a" {
	panic("Pair set error")
}

p.Clear()
if p.First != 0 || p.Second != "" {
    panic("Pair clear error")
}

p2 := Util.Pair[int, string]{}
p2.Set(2, "b")
Util.ExchangePairs(&p, &p2)

if p.First != 2 || p.Second != "b" || p2.First != 0 || p2.Second != "" {
    panic("Pair set error")
}

p = Util.MakePair[int, string](3, "d")
if p.First != 3 || p.Second != "d" {
    panic("Make error")
}
```