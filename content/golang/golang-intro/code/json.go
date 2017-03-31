package main
import "encoding/json"
import "fmt"
type Response1 struct {Page   int; Fruits []string}
type Response2 struct {
    Page   int      `json:"page"`
    Fruits []string `json:"fruits"`
}
func main() {
    slcD := []string{"apple", "peach", "pear"}
    slcB, _ := json.Marshal(slcD)
    fmt.Println(string(slcB))
    mapD := map[string]int{"apple": 5, "lettuce": 7}
    mapB, _ := json.Marshal(mapD)
    fmt.Println(string(mapB))

    res1D := &Response1{
        Page:   1,
        Fruits: []string{"apple", "peach", "pear"}}
    res1B, _ := json.Marshal(res1D)
    fmt.Println(string(res1B))
}
