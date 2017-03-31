package main
import "fmt"

func main() {
    s := make([]string, 3)
    fmt.Println("emp:", s)

    s[0], s[1], s[2] = "a", "b", "c"
    fmt.Println("set:", s, "get", s[2])
    fmt.Println("len:", len(s))
    s = append(s, "d")
    c := make([]string, len(s))
    copy(c, s)

    twoD := make([][]int, 3)
    for i := 0; i < 3; i++ {
        innerLen := i + 1
        twoD[i] = make([]int, innerLen)
        for j := 0; j < innerLen; j++ {
            twoD[i][j] = i + j
        }
    }
    fmt.Println("2d: ", twoD)
}
