Package regexp中的正则表达式默认是大小写敏感、单行匹配的，可以通过设置flag来改变默认行为，使用分组的语法设置flag：

1. (?flags)       set flags within current group; non-capturing
2. (?flags:re)    set flags during re; non-capturing

注意，只是使用分组的语法设置flag，并没有引入新的分组(non-capturing)；

第一种方案：为所在的正则表达式设置flag；
第二种方案：为re正则表达式设置flag；

Flag syntax is xyz (set) or -xyz (clear) or xy-z (set xy, clear z). The flags are:

i              case-insensitive (default false) // 匹配时忽略大小写
m              multi-line mode: ^ and $ match begin/end line in addition to begin/end text (default false) // 将^和$匹配字符串(不是字符串中的word)的开始和结束，调整为匹配各行首和行尾；
s              let . match \n (default false) // . 匹配多行；
U              ungreedy: swap meaning of x* and x*?, x+ and x+?, etc (default false) // 非greddy：交换x*和x*?的含义，即*表示非greedy，*?表示贪婪；

示例：
package main

import (
	"fmt"
	"regexp"
)

func main() {
	const s = `
foo
foo 2
bar
foo`

	r := regexp.MustCompile(`(foo).*`) // 默认.不能匹配\n，故(foo).*每次只能匹配一行；
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1)) // 返回每行的匹配情况，结果类型[][]string{}， 每一项的[]string{}的第一个元素为匹配整个正则的内容，第二个元素为sub group1;

	r = regexp.MustCompile(`(?s)(foo).*`) // 指定flag s，(?s)表示对整个正则有效；.能匹配\n，故(foo).*能够匹配整个多个字符串；注意(?s)不引入分组, 所以结果[][]string{}的每一项[]string只有两个元素；
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))

	r = regexp.MustCompile(`(?s)(foo).*`) 
	fmt.Printf("%#v\n", r.FindStringSubmatch(s)) // 匹配整个字符串
	
	r = regexp.MustCompile(`^.*foo.*$`) // ^...$表示字符串的开始和结束, .不匹配\n，故无匹配项；
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))

	r = regexp.MustCompile(`^\n.*\n.*\n.*\n.*\n.*$`) // 手动指定\n
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))

    r = regexp.MustCompile(`(?s)^.*foo.*$`) // 指定flag s, .能匹配\n, 故能匹配整个字符串；
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))

	r = regexp.MustCompile(`(?m:^foo$)`)    // (?m:^re$) 指定^re$匹配行首到行尾，而不是字符串的开始和结束，故匹配只有foo的行；
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))

	r = regexp.MustCompile(`.*`) // .*匹配每一行
	fmt.Printf("%#v\n", r.FindAllStringSubmatch(s, -1))
}

输出：
[][]string{[]string{"foo", "foo"}, []string{"foo 2", "foo"}, []string{"foo", "foo"}}
[][]string{[]string{"foo\nfoo 2\nbar\nfoo", "foo"}}
[]string{"foo1\nfoo\nfoo 2\nbar\nfoo", "foo"}
[][]string(nil)
[][]string{[]string{"\nfoo1\nfoo\nfoo 2\nbar\nfoo"}}
[][]string{[]string{"\nfoo1\nfoo\nfoo 2\nbar\nfoo"}}
[][]string{[]string{"foo"}, []string{"foo"}}
[][]string{[]string{""}, []string{"foo"}, []string{"foo 2"}, []string{"bar"}, []string{"foo"}}