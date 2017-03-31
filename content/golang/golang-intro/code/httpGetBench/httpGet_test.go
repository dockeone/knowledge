package httpServerBench

import ("net/http"; "io/ioutil"; "testing")

func HttpGet(){
    resp, _ := http.Get("http://127.0.0.1:8000")
    defer resp.Body.Close()
    ioutil.ReadAll(resp.Body)
}

func benchmarkHttpGet(i int, b *testing.B) {
        for n := 0; n < b.N; n++ {
               HttpGet() 
        }
}

func BenchmarkHttpGet1(b *testing.B)  { benchmarkHttpGet(1, b) }
func BenchmarkHttpGet2(b *testing.B)  { benchmarkHttpGet(2, b) }
func BenchmarkHttpGet3(b *testing.B)  { benchmarkHttpGet(3, b) }
func BenchmarkHttpGet10(b *testing.B) { benchmarkHttpGet(10, b) }
func BenchmarkHttpGet20(b *testing.B) { benchmarkHttpGet(20, b) }
func BenchmarkHttpGet40(b *testing.B) { benchmarkHttpGet(40, b) }
