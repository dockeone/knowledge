<!-- toc -->

# The Go Programming Language Specification

# Introduction
Go是一个通用的系统编程语言。它支持强类型、垃圾回收和并发编程。Go程序由非常有效地管理依赖的package组织。当前使用传统的编译、链接方式生成可执行文件。

# Source code representation
源代码是UTF-8编码的文本文件。

## Characters
下列terms用于表示Unicode字符类(character classes)：

	newline        = /* the Unicode code point U+000A */ .
	unicode_char   = /* an arbitrary Unicode code point except newline */ .
	unicode_letter = /* a Unicode code point classified as "Letter" */ .
	unicode_digit  = /* a Unicode code point classified as "Decimal Digit" */ .

`unicode_char`的范围比`unicode_letter`大。

下划线被当做一个字母。

	letter        = unicode_letter | "_" .

# Lexical elements

## Comments
两种注释方式，不能被嵌套使用。

+ 行注释：从字符序列`//`开始直到行尾，行注释被当做一个换行。
+ 一般注释：从字符序列`/*`开始直到`*/`结束。

包含换行的注释被当做一个换行，否则当做一个空格。

## Tokens
Tokens形成了Go语言的词汇。有四种类型的Tokens：标识符(identifiers)、关键字(keywords)、操作符和分隔符(operators and delimiters)和字面量(literals)。

各种空格(`空格符、\f、\t、\r、\n`)在编译时会被忽略，但是用于分隔token的空格是不会被忽略的，否则多个tokens会错误地被当做一个更长的token；

将源文件划分token时，总使用下一个最长的token；

编译器可能在换行或文件的末尾插入一个分号。

由于编译器会忽略空格，所以Go源文件的格式是**自由的**。

编译器读取源文件，去掉注释后，划分为一系列Tokens，然后解释token的含义,如表达式、声明、语句等；

## Semicolons
正式的语法使用分号来终止(*terminate*)输出， 但在下面情况下，编译器自动添加分号：

1. 当前行非空且最后一个token是：

	+ an identifier
	+ an integer, floating-point, imaginary, rune, or string literal
	+ one of the keywords break, continue, fallthrough, or return
	+ one of the operators and delimiters ++, --, ), ], or }

2. To allow complex statements to occupy a single line, a semicolon may be omitted before a closing ")" or "}".

## Identifiers
标识符用于语命名程序中的实体(*entities*)，例如变量和类型。标识符是字母和数字的序列，第一个字符必须是**字母**（字母包含下划线`_`）:

	identifier = letter { letter | unicode_digit } .

Go预定义了一些标识符(标识符和关键字是有区别的，后者不能被重新绑定)。

Types:
	bool byte complex64 complex128 error float32 float64
	int int8 int16 int32 int64 rune string
	uint uint8 uint16 uint32 uint64 uintptr

Constants:
	true false iota

Zero value:
	nil

Functions:
	append cap close complex copy delete imag len
	make new panic print println real recover

## Keywords
下列为保留的标识符，它们的含义不能重新定义：

	break        default      func         interface    select
	case         defer        go           map          struct
	chan         else         goto         package      switch
	const        fallthrough  if           range        type
	continue     for          import       return       var

有些预定义的类型和常量值，虽然不是关键字，但是也不建议重复定义：

## Operators and Delimiters
下列字符序列被当做运算符、分隔符和其它特殊tokens：

	+    &     +=    &=     &&    ==    !=    (    )
	-    |     -=    |=     ||    <     <=    [    ]
	*    ^     *=    ^=     <-    >     >=    {    }
	/    <<    /=    <<=    ++    =     :=    ,    ;
	%    >>    %=    >>=    --    !     ...   .    :
    &^          &^=


## Integer literals
整形字面量由数字序列组成，代表了整形常量。数字序列默认为十进制，可以在前面添加可选前缀来改变进制：0表示八进制、0x或0X表示十六进制。

注意：没有二进制字面量。

## Floating-point literals
浮点字面量是十进制表示的浮点常量，由整数部分、小数点、小数部分和指数部分组成。

	float_lit = decimals "." [ decimals ] [ exponent ] |
            decimals exponent |
            "." decimals [ exponent ] .
	decimals  = decimal_digit { decimal_digit } .
	exponent  = ( "e" | "E" ) [ "+" | "-" ] decimals .

浮点数只有十进制表示方式，无八进制、十六进制、二进制表示方式；

浮点字面量示例：
	0. // 省略小数和指数
	72.40 // 省略指数
	072.40  // == 72.40 // 注意：以0开头的整数如072会被当做八进制；
	2.71828
	1.e+0
	6.67428e-11
	1E6
	.25
	.12345E+5

## Imaginary literals

	imaginary_lit = (decimals | float_lit) "i" .

## Rune literals
rune字面量代表一个Unicode code point的`整型值`；rune是int32的别名(alias)，两者可以通用。

rune字面量用**单引号**引住，单引号内可以是除了单引号和换行符外的任意字符：单个字符代表它本身，以反斜杠开头的多个字符有特殊含义。

	rune_lit         = "'" ( unicode_value | byte_value ) "'" .
	unicode_value    = unicode_char | little_u_value | big_u_value | escaped_char .
	byte_value       = octal_byte_value | hex_byte_value .
	octal_byte_value = `\` octal_digit octal_digit octal_digit .
	hex_byte_value   = `\` "x" hex_digit hex_digit .
	little_u_value   = `\` "u" hex_digit hex_digit hex_digit hex_digit .
	big_u_value      = `\` "U" hex_digit hex_digit hex_digit hex_digit
                           hex_digit hex_digit hex_digit hex_digit .
	escaped_char     = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` ) .

rune中转意字符具有特殊含义，其它转移字符是非法的：

## String literals
String 字面量代表字符串常量，有两种形式：原生字符串字面量(raw string literals)和被解释的字符串字面量(interpreted string literals)。

原生字符串字面量由反引号包裹，在引号内部，除了反引号外，其它任何字符都有效，转意字符没有特殊含义，可以包含回车。换行符(`\r`)将被忽略。

解释性字符串字面量由双引号包裹，在引号内部不能包含回车，转意字符有特殊含义。

	string_lit             = raw_string_lit | interpreted_string_lit .
	raw_string_lit         = "`" { unicode_char | newline } "`" .
	interpreted_string_lit = `"` { unicode_value | byte_value } `"` .  #unicode_value包含unicode_char

# Constants
6种常量：boolen、rune、interger、float、complex和string，其中 rune、integer、floating、complex被称为数值常量。
数值常量代表精确值，不会溢出；组合类型如struct、map、channel、array不是常量；

注意常量和字面量的区别！

预定义的true和false代表boolean常量，预定义的标识符iota代表整形常量；

常量可能typed或untyped:

+ 字面量常量、true、false、iota和只包含无类型的常量操作数的常量表达式是untyped常量；
+ 常量可以被赋予类型：显式的常量声明或转换、隐式的变量声明或赋值(结果是一个变量)；如果常量值不能代表该类型的值(如将3.01转换为int型的3), 则会发生错误；

无类型的常量具有缺省类型，在需要typed value的场合，常量将隐式的转换为相应的缺省类型；如 i := 0，常量的缺省类型是：bool、rune、int、float64、complex128和string；

# Variables
变量代表保存value的存储位置，变量的类型(type)决定了变量的取值范围；

new内置函数、获取组合字面量的地址会在runtime时分配存储空间，结果可以作为变量使用；

结构化变量如array、slice、struct的成员有可能是可寻址的(匿名结构化变量是不可寻址的)，可作为变量使用(如表达式中使用，或赋值语句左侧使用)；

注意，map的value是不可寻址的，不能作为变量在赋值语句左侧使用，但是支持m["a"]++运算符；

变量的静态类型(简称类型)在变量声明、new、组合字面量、结构化变量的元素时指定；
接口类型的变量还具有动态类型，该类型是在给该接口类型变量赋值的值类型；

	var x interface{}  // x is nil and has static type interface{}
	var v *T           // v has value nil, static type *T
	x = 42             // x has value 42 and dynamic type int
	x = v              // x has value (*T)(nil) and dynamic type *T

变量如果没有被赋值，则它的value是相应类型的zero value；

变量可以在声明时赋值，也可以在声明后赋值，注意slice、map、channel声明后未赋值前的zero value是nil，需要使用make()来进行赋值；

# Types
类型决定取值范围和可用的操作；类型有named和unamed之分：

+ named类型用type name来指定，包括预定义的boolean， numeric、string类型；
+ unamed类型用类型字面量(type literal)来指定，类型字面量是用已有类型来组成的，包括 array、struct、pointer、function、interface、slice、map、channel；

        Type      = TypeName | TypeLit | "(" Type ")" .
        TypeName  = identifier | QualifiedIdent .
        TypeLit   = ArrayType | StructType | PointerType | FunctionType | InterfaceType | SliceType | MapType | ChannelType .

每个type T都有一个underlying type：
+ 如果T是预定义的named类型或type literal，则underlying type是T本身；
+ 否则，T的underlying type是T指向类型的underlying type，这种指向具有递归性直到符合上条情况；

   type T1 string
   type T2 T1
   type T3 []T1
   type T4 T3

The underlying type of string, T1, and T2 is string. The underlying type of []T1, T3, and T4 is []T1.

类型T的操作特性(运算符等)取决于underlying type，underlying type也影响赋值；

## Method sets
类型可能有关联的 method set：
+ 接口类型的method set是它本身；
+ 非接口类型T的方法集是用receiver type T声明的methods；
+ 非接口类型T的指针类型 *T 对应的方法集是用receiver type T或\*T声明的methods；
+ 所有type均实现了interface {};

类型的方法集决定了该类型实现的接口以及用相应receiver type可以调用的methods；

方法集一般用在接口类型，例如函数的参数是interface T时，可以传入的值 X 或 \*X 的方法集必须实现T；

## Boolean types
两个预定义的常量值 true和false，预定义类型是bool；

## Numeric types
使用二进制补码表示，最高位为符号位。
+ 对于有符号类型位数n，有效范围为：-2^(n-1) ~ 2^(n-1)-1, -2^(n-1)为-0;
+ 对于无符号类型位数n，有效范围为 0 ~ 2^n；

    byte        alias for uint8
    rune        alias for int32

实现相关的类型:

    uint     either 32 or 64 bits
    int      same size as uint
    uintptr  an unsigned integer large enough to store the uninterpreted bits of a pointer value

所有numeric types是不同的，编译器不会隐式转换，在混合运算时，必须显式转换为相同类型后运算；但byte是unit8的alias，rune是int32的alias，相互通用；

## String types
string是bytes序列，不可变；不能使用cap()函数（array也不能使用)； &s[i] 操作非法；

## Array types
array是单一类型value的序列，序列的数量称为长度；长度是array type的一部分，执行结果必须是可以用int代表的非负常量；
array的元素是可以寻址的；

## Sliece types
未初始化的slice是nil；
slice用于描述underlying array的连续段(contiguous segment)，并提供访问该连续段array的方式；
slice的长度是可变的，可以通过reslice扩充其长度(只要不超过underlying arry的长度，即slice的capiticy即可)；
slice的元素是可以寻址的；

未初始化赋值的slice值为nil；可以使用make函数来创建并初始化一个slice，capacity参数是可选的(length是必须参数)：

    make([]T, length, capacity)

下面两个语句等效：

    make([]int, 50, 100)
    new([100]int)[0:50]

对于多维的slice，内层的slice长度是可变的且必须要单独initialize;

无元素的slice字面量的length和capacity是0，不能index成员，如：
v =: []int{} // v != nil，但是length和capacity是0；
v[0] = 1 // panic

nil的slice支持append函数；

## Struct types
只有type没有name的filed声明被称为匿名field或嵌入的field；可以嵌入类型T或非接口类型的*T，T本身不能为指针类型；
嵌入的unqualified type name 被当做filed name：

    // A struct with four anonymous fields of type T1, *T2, P.T3 and *P.T4
    struct {
	    T1        // field name is T1
	    *T2       // field name is T2
	    P.T3      // field name is T3, 注意没有P；
	    *P.T4     // field name is T4
	    x, y int  // field names are x and y
    }

struct x的匿名field的field或方法f可以提升(promoted)为x的field或方法；但是提升的filed不能在struct x的组合字面量中使用，而应该将匿名field作为整体赋值；

struct 类型S 和 类型 T，S包含的提升方法规则如下：
1. 如果S包含匿名成员T，S 和 *S 的方法集包含提升的T的方法； *S的方法集还包括提升的\*T的方法；
2. 如果S包含匿名成员*T, S 和 *S 的方法集包含T和\*T的方法集；

示例：
//Nodes 保存注册的Node信息，使用Mutex锁对map的并发读写进行保护
type Nodes struct {
	*sync.RWMutex
	Nodes map[string]Node //string: pcap | slave
	_ float32  // padding
}
var nodes Nodes;

则nodes值包含接受类型为sync.RWMutex或*sync.RWMutex的方法集；

## Pointer types
未初始化的指针类型值是nil；
指针类似是可比较的，但是不能运算；

## Funciton types
未初始化的函数类型值是nil；

函数不能嵌套定义；但是可以在函数内部定义并使用匿名函数；

函数**最后一个参数**类型前可以加...表示可变参数，在调用时，可以传入0个或多个相应类型的值，该参数会被声明为对应类型的slice(但是未初始化，地址和长度均为0，不能index，但可以append)

## Interface typs
接口类型用于指定方法集；支持嵌套其它接口，但不支持嵌入接口本身，如果嵌套引起循环依赖，会出错；

## Map types
未初始化的map类型值是nil，nil map不支持赋值，但是可以获取结果；
可以使用make函数定义初始化的map：

    make(map[string]int)
    make(map[string]int, 100)

key type必须是可比较的，即支持==和!=操作符，所以function、map、slice类型值不能用作key；
m["index"]的结果是不可寻址的(因为map动态增长后，以前的地址就可能会无效)，但支持m[k]++运算符；使用delete函数删除map的元素；
如果元素不存在，m[index]返回的结果是元素的zero value，可以使用多重赋值来判断是否存在该元素；

## Channel types
未初始化的channel类型值是nil；可以使用make函数定义初始化的channel:

    make(chan int, 100)

第二个参数指定channel的capacity，同时也是channel buffer大小；如果值为0或未指定，则channel是unbufferd，当sender和receiver都ready的时候才能通信；
否则channel是buffered，当buffer没有full或empty之前，sender和receiver都可以通信；

nil channel不能用于发送或接收，否则引起runtime panic；

内置函数close关闭channel，读取关闭的channel返回对应元素类型的零值，可以使用channel接收操作符<-的多重赋值形式来判断channel是否close。

len(chan T)返回channel buffer中queued的元素数目；cap(chan T)返回channel buffer的容量(传给make函数的参数值)；

channel的方向运算符<-是向左最长关联的， 必要的时候需要用括号：

    chan<- chan int    // same as chan<- (chan int)
    chan<- <-chan int  // same as chan<- (<-chan int)
    <-chan <-chan int  // same as <-chan (<-chan int)
    chan (<-chan int)

chan可发送或接收的数据类型是没有限制的！

# Properties of types and values

## Type identify
Go是强类型语言，两个type相同(identical)或不相同(different)，类型影响到Assignability；
+ 使用相同type声明的两个命名types相同；
+ 命名和非命名的types不相同；
+ 使用相同type literal定义的两个非命名的types相同：
  - 数组元素类型和长度相同；
  - slice元素类型相同；
  - struct的field名称、类型、顺序、tags相同；如果有Lower-case的field且来源于不同packages，则两个struct不同；
  - 相同base类型的指针；
  - 相同签名的函数，忽略参数名；
  - 相同方法集合、相同函数类型的接口，函数的顺序无关；
  - 相同key和value类型的map；
  - 相同元素类型和方向的channel；

    type (
        T0 []string
        T1 []string
        T2 struct{ a, b int }
        T3 struct{ a, c int }
        T4 func(int, float64) *T0
        T5 func(x int, y float64) *[]string
    )

    these types are identical:

    T0 and T0
    []int and []int
    struct{ a, b *T5 } and struct{ a, b *T5 }
    func(x int, y float64) *[]string and func(int, float64) (result *[]string)

    T0 and T1 are different because they are named types with distinct declarations;
    func(int, float64) *T0 and func(x int, y float64) *[]string are different because T0 is different from []string.

## Assignability

在下列情况下，值x可以赋值给T类型的变量：
+ x的类型和T一致；
+ x的类型V和T具有相同的underlying types，且V和T至少有一个是非命名类型(如字面量值或字面量类型)；
+ T是接口类型，x实现T；
+ x是一个双向channel value，T是channel类型，x的类型V和T有相同的元素类型且V和T至少有一个是非命名类型；
+ x是预定义标识符nil，T是指针、函数、slice、map、channel或接口类型；
+ x是无类型的常量，且可被T类型的value所代表(representable); 如: var i int32 = 10.0
  这里的representable意味着，无类型数值常量(int或float)可以赋值给有类型数值**变量**，只要变量类型能容纳下对应的常量值即可；
  const i = int(10.0) 是错误的！

# Blocks

Block是大括号包含的声明(Declaration)和语句(Stmt)列表(可以为空)；

    Block = "{" StatementList "}" .
    StatementList = { Statement ";" } .

	Statement =
		Declaration | LabeledStmt | SimpleStmt |
		GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
		FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
		DeferStmt .

	SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .

块(Block)是大括号包含的语句列表(StatementList)组成，语句包含声明(Declaration)和其它类型的语句(空语句、表达式语句、赋值语句、短变量声明等)，语句由分号结束；

除了显式的大括号定义的block外，下列情况隐式生成block：
1. universe block包含所有go源文件；
2. 每个package有一个package block，包含该package的所有源文件；
3. 每个file有一个file block，包含该文件的所有内容；
4. 各if、for、switch有各自的隐式block；
5. switch、slect语句的各clause有各自的隐式block；

注意：4和5意味着在if、for等语句定义的变量，只在该语句块中有效，语句结束后变量就被GC回收; 

如果在隐式block中使用:=运算符，则定义新的变量；如下面的i变量：

    for i := 1; i < 10; i++ {
    }

Block影响作用域；注意：变量的作用域和生命期不是一个概念：
1. Block决定变量的作用域，作用域指的是引用范围；
2. 生命期是变量引用的对象的生成时间，生命期由GC控制，当变量的内存区域的引用计数为0时，GC将销毁对应的内存区域；

# Declarations and scope

声明(declartion)将非blank的标识符与常量、类型、变量、函数、标签和package名称绑定；
1. 程序中的所有标识符都需要声明；
2. 不允许在同一个block重复声明标识符；
3. 不允许在file和package block中重复声明同一个标识符；

blank标识符不引入新的绑定(binding)，不是声明语句，故 _ := 2或者 var _ int = 2 是错误的；

在package block，init标识符只能用于init函数声明，和blank标识符一样，不引入新的绑定，故package可以声明多个init函数；

	Declaration   = ConstDecl | TypeDecl | VarDecl .
	TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .

Go使用基于block的词法作用域，作用域决定了标识符的有效使用范围(注意，与标识符的生存期是有区别的，后者是运行时特性，由GC控制)：
1. 预定义的标识符是全局作用域；
2. Top leve声明的常量、类型、变量、函数(非方法)是package block作用域(所有函数外部，顺序无关，使用的位置可以在声明前)；
3. imported的package name是file作用域，只在当前导入的file中有效；
4. 方法接收者、函数参数、结果变量是函数body作用域；
5. 在函数体内定义的常量或者变量，从开始定义的位置到最内层block一直有效；
6. 在函数体内定义的类型，从开始定义的位置到最内层block一直有效；

外层block定义的标识符，可以在内层重新定义，内层的定义将重载外层的定义；

package子句不是声明语句，不引入任何绑定，它的作用是指定属于同一个package的源文件，同时用于后续import该package的name；

## Label scopes

Labels用于break、continue和goto语句；可以定义未使用的label，label不是block作用域，不与同名的非lable的标识符冲突；
label的作用域是定义它的函数体，不包括嵌套的函数；

## Exported identifiers

标识符可以被导出，在其它packages中使用：
1. 标识符的第一个字符是大写字母；
2. 标识符位于package level，或者是成员或方法名；

其它类型的标识符都不会被exported；

## Constant declarations

常量声明是将identifiers与**常量表达式列表**绑定；

	ConstDecl      = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
	ConstSpec      = IdentifierList [ [ Type ] "=" ExpressionList ] .

	IdentifierList = identifier { "," identifier } .
	ExpressionList = Expression { "," Expression } .

如果指定了类型，表达式必须assignable to that type；
如果忽略了类型，则常量为各表达式结果的类型；如果表达式结果是无类型常量，则声明的常量任然是无类型的：

	const Pi float64 = 3.14159265358979323846
	const zero = 0.0         // untyped floating-point constant
	const (
		size int64 = 1024
		eof        = -1  // untyped integer constant
	)
	const a, b, c = 3, 4, "foo"  // a = 3, b = 4, c = "foo", untyped integer and string constants
	const u, v float32 = 0, 3    // u = 0.0, v = 3.0

在括起来的const声明列表表达式语句中，除了第一个声明外，后续的表达式列表可以省略，这时相当于第一个类型和表达式的文本替换， 如：

const (
	Sunday = iota
	Monday  #后续的常量是前一个类型和express的重复；
	Tuesday
	Wednesday
	Thursday
	Friday
	Partyday
	numberOfDays  // this constant is not exported
)

## iota

iota通常用于const enum列表的定义中：
1. 代表持续增长的无类型整形常量；
2. const修饰符重置iota的值为0；

	const ( // iota is reset to 0
		a = 1 << iota  // a == 1
		b = 1 << iota  // b == 2
		c = 3          // c == 3  (iota is not used but still incremented)
		d = 1 << iota  // d == 8
	)

	const ( // iota is reset to 0
		u         = iota * 42  // u == 0     (untyped integer constant)
		v float64 = iota * 42  // v == 42.0  (float64 constant)
		w         = iota * 42  // w == 84    (untyped integer constant)
	)

在一个表达式内部，iota的值是相同的，因为iota只在新的const表达式中递增：
	const (
		bit0, mask0 = 1 << iota, 1<<iota - 1  // bit0 == 1, mask0 == 0，两个iota值相同；
		bit1, mask1                           // bit1 == 2, mask1 == 1
		_, _                                  // skips iota == 2
		bit3, mask3                           // bit3 == 8, mask3 == 7
	)

## Type declarations

类型声明将标识符和具有相同underlying类型的新类型绑定，所有可以用于underlying的操作都已运用于新的类型；
新类型和旧类型是不相同的；

声明的新类型不继承已有类型的任何方法；但是接口类型的方法集，组合类型的成员&方法集不变：

	// A Mutex is a data type with two methods, Lock and Unlock.
	type Mutex struct         { /* Mutex fields */ }
	func (m *Mutex) Lock()    { /* Lock implementation */ }
	func (m *Mutex) Unlock()  { /* Unlock implementation */ }

	// NewMutex has the same composition as Mutex but its method set is empty.
	type NewMutex Mutex // 新类型不继承旧类型的任何方法

	// The method set of the base type of PtrMutex remains unchanged,
	// but the method set of PtrMutex is empty.
	type PtrMutex *Mutex // 新类型不继承旧类型的任何方法

	// The method set of *PrintableMutex contains the methods
	// Lock and Unlock bound to its anonymous field Mutex.
	type PrintableMutex struct { // 新类型包含匿名field的成员和方法
		Mutex
	}

	// MyBlock is an interface type that has the same method set as Block.
	type MyBlock Block // Block是接口类型，新类型和Block具有相同的方法集；

## Variable declarations

变量声明用于创建一个或多个变量，并给各变量赋初始值；没有赋值的变量被初始化对应的zero value；

将无类型的常量赋值给变量时，会先被转换为它的default type（注意，整数的default type是int，而非int64，浮点数的default type是float64）;

nil不能给无类型的变量初始化赋值：
	var n = nil            // illegal

## Short variable declarations

短变量声明语句可能会重新声明(redeclare)位于同一个block(包括函数参数和返回值)的变量，它们的类型相同且至少有一个非空的变量；
只允许使用短变量声明的形式redeclare变量；redeclare变量不会引入新的变量，只是对原变量赋值；
只能在函数内部使用短变量声明语句；
在if、for、switch等支持子block的语句中，可以创建临时变量；

field1, offset := nextField(str, 0)
field2, offset := nextField(str, offset)  // redeclares offset
a, a := 1, 2                              // illegal: double declaration of a or no new variable if a was declared elsewhere

## Funciton declaration

如果函数签名包含返回结果参数，则函数体**必须**以终止语句(terminating statement)结束；
	func IndexRune(s string, r rune) int {
		for i, c := range s {
			if c == r {
				return i
			}
		}
		// invalid: missing return statement
	}

函数声明可以忽略body，这表示该函数在Go外部实现，如assembly routine：

func flushICache(begin, end uintptr)  // implemented externally，忽略了大括号body，而不是空body；

## Method declarations

	MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
	Receiver     = Parameters .

Receiver对应的Parameters必须是单一的非可变类型的参数，类型为T或*T, T不能为指针或接口类型；

如果函数体内不使用Receiver值，则可以忽略value标示符(只提供type)；

如果Receiver为指针类型，则可以在方法体内修改Receiver的值；

对于base type，所有non-blank的方法名称必须各不相同；如果base type是struct，则non-blank的方法名和field名称必须各不相同；

# Expression

表达式：使用运算符和函数对操作数(operands)进行运算，获取运算结果；

## Operands

Operands为表达式中的元素value，可以是字面量(literal)、非blank的标识符(代表常量、变量、函数等)、生成函数的方法表达式(method exporession)或括起来的表达式；

Operand     = Literal | OperandName | MethodExpr | "(" Expression ")" .
Literal     = BasicLit | CompositeLit | FunctionLit .
BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
OperandName = identifier | QualifiedIdent.

bool类型包含两个预定义的常量值true和false，归于上面的OperandName中；

## Composite literals

组合字面量包括structs、arrays、slices和map，每执行一次就生成新的value；
使用大括号分割的元素列表组成，每个元素前面可以有个key；

CompositeLit  = LiteralType LiteralValue .
LiteralType   = StructType | ArrayType | "[" "..." "]" ElementType |
                SliceType | MapType | TypeName .
LiteralValue  = "{" [ ElementList [ "," ] ] "}" .
ElementList   = KeyedElement { "," KeyedElement } .
KeyedElement  = [ Key ":" ] Element .
Key           = FieldName | Expression | LiteralValue .
FieldName     = identifier .
Element       = Expression | LiteralValue .

对于struct字面量而言，key为filedname(即标示符identify)；
对于array和slice字面量而言，key为index；对于map而言为key；index和key可以是表达式或者字面量值；

在组合字面量中，多次对同一个field name或constant key value赋值是错误的；

对于struct字面量而言：
1. key必须是struct的filed；
2. 如果字面量不包含key，则必须按顺序列出所有filed的值；
3. 如果一个元素有key，其它元素也必须用key指定；
4. 如果用key指定元素，则不必列出所有key，为列出的key值为对应的zero值；
5. 如果忽略struct的元素列表，则每个元素为对应的zero值；
6. 不能给其它package中非export的field赋值；

对于slice和array字面量而言：
1. 每个元素关联一个整型index，表示它在array中的位置；
2. 如果一个元素指定了key index，则它必须是整型表达式(不一定是常量)；
3. 如果元素没有指定index，则它的index为前一个元素的index值+1，第一个元素的index值为0；

组合字面量是可寻址的，寻址运算符会新初始化一个变量，返回它的地址；

组合字面量支持如下的简写形式：

	[...]Point{{1.5, -3.5}, {0, 0}}     // same as [...]Point{Point{1.5, -3.5}, Point{0, 0}}
	[][]int{{1, 2, 3}, {4, 5}}          // same as [][]int{[]int{1, 2, 3}, []int{4, 5}}
	[][]Point{{{0, 1}, {1, 2}}}         // same as [][]Point{[]Point{Point{0, 1}, Point{1, 2}}}
	map[string]Point{"orig": {0, 0}}    // same as map[string]Point{"orig": Point{0, 0}}
	[...]*Point{{1.5, -3.5}, {0, 0}}    // same as [...]*Point{&Point{1.5, -3.5}, &Point{0, 0}}，自动获取地址；
	map[Point]string{{0, 0}: "orig"}    // same as map[Point]string{Point{0, 0}: "orig"}

如果在if、for、switch的子句中使用TypeName类型的字面量，则需要使用括号，防止引起歧义：
if x == (T{a,b,c}[i]) { … } //如果不加括号则解释为: 将 x 和 T进行比较，函数体为 {a, b, c}, 后续出现语法错误；
if (x == T{a,b,c}[i]) { … }

// the array [10]float32{-1, 0, 0, 0, -0.1, -0.1, 0, 0, 0, -1}
filter := [10]float32{-1, 4: -0.1, -0.1, 9: -1}

## Function literals

函数字面量代表匿名函数；匿名函数具有闭包(closures)特性：可以引用外围函数定义的变量；

	f := func(x, y int) int { return x + y }
	func(ch chan int) { ch <- ACK }(replyChan)

注意：下面使用匿名函数的方式是错误的：
	for i, file = range os.Args[1:] {
		wg.Add(1)
		go func() {
				compress(file)
				wg.Done()
		}() 
	} 
因为go是延迟执行的，函数内的file引用的是执行时的file值，而不是循环时的file值；解决方法是将file作为参数传给goroutine函数，或者调用goroutine前
将file赋值给临时变量，在groutine内部使用该临时变量；

# Primary exprssion

主表达式是单目或双目运算符的操作数(operands)，故优先级比运算符高，如表达式：(*t.T0).x ：
t.T0 是Primary expression，*t.T0 是对Primay expression使用\*单目运算符，(\*t.TO)作为一个整体的Operand，然后(\*t.T0).x是Primay expression

	PrimaryExpr =
		Operand |
		Conversion |
		PrimaryExpr Selector |
		PrimaryExpr Index |
		PrimaryExpr Slice |
		PrimaryExpr TypeAssertion |
		PrimaryExpr Arguments . // 函数调用

	Selector       = "." identifier .
	Index          = "[" Expression "]" .
	Slice          = "[" [ Expression ] ":" [ Expression ] "]" |
					"[" [ Expression ] ":" Expression ":" Expression "]" .
	TypeAssertion  = "." "(" Type ")" .
	Arguments      = "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .

选择、索引、slice、类型断言、函数调用都是PrimaryExpr，优先级比单目和双目运算符高；

## Selectors

x不是package name，x.f 代表值x(或*x)的成员或方法f；

选择符f可能代表类型T的成员或方法，也可能是类型T嵌套成员的成员或方法，嵌套的次数称为深度；
1. x是T或*T类型的值，且T不是指针或接口类型，x.f代表T最浅深度的成员或方法，如果同一深度有多个相同名成员或方法，则非法； <- 不考虑T或*T类型的方法集，因为T不是接口类型；
2. 如果x是接口类型I的值，x.f代表接口动态类型对应的成员或方法，f必须位于I的**方法集**中；
3. 1和2条件的特殊情况：如果x是命名的指针类型，且(*x).f是一个有效的成员(filed，非method)，x.f和（\*x).f等效;
4. 其它情况下，x.f是非法的；
5. 如果x是指针类型且值为nil，x.f如果代表struct field，则读取或赋值 x.f将引起runtime异常；
6. 如果x是接口类型且值为nil，x.f如果代表方法，则调用x.f时将引起runtime异常；

对于条件1，如果T不是接口类型或指针类型：
1. 对于T类型的值x，x可以调用在T或*T上定义的方法； 这意味着T类型值x可以调用receiver类似是*T或T的方法；
2. 对于*T类型的值x，x可以调用在T或*T上定义的方法；


示例，如下声明语句：
	type T0 struct {
		x int
	}

	func (*T0) M0()

	type T1 struct {
		y int
	}

	func (T1) M1()

	type T2 struct {
		z int
		T1
		*T0 //命令的指针类型
	}

	func (*T2) M2()

	type Q *T2 //命名的指针类型

	var t T2    
	var p *T2   
	var q Q = p

	one may write:

	t.z          // t.z
	t.y          // t.T1.y
	t.x          // (\*t.T0).x   t.T0是命名的指针类型且(*t.T0).x是有效的选择符，所以可以使用t.x缩写形式；

	p.z          // (\*p).z  p是命名的指针类型\*T2，且 (*p).z有效，故可以使用p.z缩写形式；
	p.y          // (*p).T1.y
	p.x          // (*(*p).T0).x

	q.x          // (*(*q).T0).x        (*q).x is a valid field selector

	p.M0()       // ((*p).T0).M0()      M0 expects *T0 receiver
	p.M1()       // ((*p).T1).M1()      M1 expects T1 receiver
	p.M2()       // p.M2()              M2 expects *T2 receiver
	t.M2()       // (&t).M2()           M2 expects *T2 receiver, see section on Calls  注意这种有效形式，自动传入t的地址！

	but the following is invalid:

	q.M0()       // (*q).M0 is valid but not a field selector

## Method expressions

如果M是类型T方法集中一个方法，则T.M可以像普通函数一样使用，只不过第一个参数为方法集中M的接收者；
	type T struct {
		a int
	}
	func (tv  T) Mv(a int) int         { return 0 }  // value receiver
	func (tp *T) Mp(f float32) float32 { return 1 }  // pointer receiver

	var t T

	T.Mv 生成一个函数，等效于：

	func(tv T, a int) int  //第一个参数为T方法集中Mv的接收者类型；

	其它调用方式：
	t.Mv(7)
	T.Mv(t, 7) // 方法表达式调用
	(T).Mv(t, 7)
	f1 := T.Mv; f1(t, 7)
	f2 := (T).Mv; f2(t, 7)

	类似的：
	(*T).Mp //注意，T.Mp是无效的，因为T类型的方法集中不包含Mp方法；
	等效于：
	func(tp *T, f float32) float32

	由于*T的方法集中包含T作为参数的方法Mv，故下面的表达式也是有效的：
	(*T).Mv //有效！
	等效于：
	func(tv *T, a int) int   //go根据传入的地址tv创建一个变量，函数不会修改通过地址传进来的参数；

## Method values

如果值t的静态类型T有方法M，则t.M被称为方法值，它是一个可以调用的函数；

	type T struct {
		a int
	}
	func (tv  T) Mv(a int) int         { return 0 }  // value receiver
	func (tp *T) Mp(f float32) float32 { return 1 }  // pointer receiver

	var t T
	var pt *T
	func makeT() T

	t.Mv 生成一个函数：func(int) int

	f := t.Mv; f(7)   // like t.Mv(7)
	f := pt.Mp; f(7)  // like pt.Mp(7)
	f := pt.Mv; f(7)  // like (*pt).Mv(7) // 注意：自动转换；
	f := t.Mp; f(7)   // like (&t).Mp(7)  // 注意：自动转换；
	f := makeT().Mp   // invalid: result of makeT() is not addressable

	var i interface { M(int) } = myVal
	f := i.M; f(7)  // like i.M(7)

## Index expressions

主表达式a[x]被称为索引表达式；a可以是数组、**数组的指针**、slice、字符串和map；

+ 如果a不是map类型： x必须是整型或者无类型数字，必须位于0 <= x < len(a)，否则panic；
+ 如果a是array类型的指针: a[x]是等效于(*a)[x]
+ 如果a是string类型：a[x]结果是非常量(non-constant byte value)byte类型，不可以给a[x]赋值；a[x]是不可以寻址的；
+ 如果M是map类型且为nil(未初始化的map)或者不包含x对应的元素，则a[x]返回对应的zero value；如果向nil的map插入值，则panic；

其它情况下 a[x]是非法的。例如a是指向slice、map、字符串的指针，则a[x]是非法的。

## Slice expressions

slice表达式产生一个子字符串或者slice，可以应用于string、数组、数组指针、slice；

### simple slice expressions

a[low : high]

high >= low >= 0; index不能是负值，不支持倒序索引；

a[2:]  // same as a[2 : len(a)]
a[:3]  // same as a[0 : 3]
a[:]   // same as a[0 : len(a)]

如果a是array的指针，则 a[low : high]等效于 (*a)[low : high]；
如果a是array或strings，indices应该位于 0 <= low <= high <= len(a)，否则panic；
对于slice而言，上边界取决于cap(a)的值，而非length；

除了无类型的strings，对string和slice进行slide操作后，结果还是非常量的string或slice类型值；

对array进行slice操作时，该array必须是可寻址的，结果是slice类型；

### full slice expressions

只对array、array指针、slice，不包含string，主表达式：

	a[low : high : max]

结果为相应类型的slice，容量为max-low; 只有low可以省略，默认为0；

如果是array，它必须是可寻址的；

0 <= low <= high <= max <= cap(a), 否则panic；

## Type assertions

x.(T)：表达式x的结果**必须是接口类型且不能为nil**；如果aeesertion的结果为fase，则会引起run-time panic;
解决办法是使用多变量赋值的语句判断：
	var v, ok T1 = x.(T)

## Calls

函数的形参必须是single-valued表达式，且assignable to 函数的参数类型；在调用函数前，先执行传递的表达式参数；
调用nil函数将引起run-time panic;
如果函数的返回参数数目和类型与另一个函数的匹配，则可以将函数作为因一个函数的参数：f(g(parameters_of_g))；
如果f的最后一个参数v是...类型，则它的内容是g函数返回值各自赋值非f后剩余的结果；

如果x的方法集中包含m，则x.m()合法；如果x是可寻址的且&x的方法集中包含m，则x.m()等效于(&x).m();

### Passing arguments to ... Parameters

如果函数最后一个参数p类型是 ...T，则在函数内部p的类型等效于[]T，如果函数调用时没有给p传入参数，则p的值为nil；

注意：只有当函数最后一个参数是可变类型且传入的slice参数类型与之一致时，才能使用 s... 的形式将整个slice传入函数；

## Operators

操作符将操作数组合为表达式；

Expression = UnaryExpr | Expression binary_op Expression .
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .

binary_op  = "||" | "&&" | rel_op | add_op | mul_op .
rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
add_op     = "+" | "-" | "|" | "^" .
mul_op     = "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .

unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .

对于双目运算符，操作数类型必须一致，例外情况是移位运算符或操作数是无类型的常量；

除了移位运算，如果一个操作数是无类型常量，另一个操作数不是，则无类型常量会被先转换为另一个操作数的类型；

## Operator precedence

单目运算符具有最高优先级； ++和--是语句，不是表达式，所以不在运算符优先级考虑的范围内，所以 *p++等效于(*p)++;

运算符有5级优先级：
Precedence    Operator
    5             *  /  %  <<  >>  &  &^    // 乘性算术运算符(包括位运算）
    4             +  -  |  ^				// 加性算术运算符
    3             ==  !=  <  <=  >  >=      // 关系
    2             &&					    // 逻辑
    1             ||

1. 左移和右移n位，分别相当于乘以2^n或除以2^n，所以<<和>>属于乘性运算符，和算术的乘、除优先级一致；
2. ^ 按位异或； &^ 按位同或；

## Arithmetic operators

算术运算符适用于数值，结果为第一操作数的类型；
+ - * / 适用于整形、浮点型和虚数类型，+ 还适用于string类型；位运算只适用于整形；

### Integer operators

 q = x / y and remainder r = x % y 满足：
 	x = q*y + r  and  |r| < |y|

如果除数是常量，则不能为0，否则会造成run-time panic；（如果除数不是constant，则可以为0，结果可能为math.Infinte，math.NaN等）；

余数的符号与x、y的符号无关，需要根据上面的表达式确定；

## Integer overflow

+ 对于无符号整形，算术运算的结果是 module 2^n, n为无符号类型的宽度(位数)，如果结果超过了2^n，则超过的(溢出的)部分将被丢弃，程序需要能处理“wrap around"的情况；
+ 对于有符号类型，算术运算的结果是 module 2^(n-1)，如果超过了该值，则可能溢出到符号位，正值可能变为负值，负值也可能变为正值；

	var u uint8 = 255
	fmt.Println(u, u+1, u*u) // "255 0 1"
	// u +1 或 u*u 的结果溢出，结果变小(wrap-around)

	var i int8 = 127
	fmt.Println(i, i+1, i*i) // "127 -128 1"
	// i+1的结果：1000 0000，1溢出到符号位，故结果为 -128;

overflow时不会产生异常；

## String concatenation

字符串相加或+=的结果为新的字符串；

## Comparison operators

比较运算符用于比较两个操作数，结果为无类型的boolean;
比较的两个操作数，必须满足：
1. 第一个操作数可以赋值给(assigned)第二个操作数；
2. 或者反之；

支持 == 和 != 的操作数是可比较的(comparable); 支持  <、<=、>、>= 的操作数是有序的(ordered);

只有整型、浮点型、字符串类型值是有序的；

对于同时指向zero-size值的指针变量，可能相等，也可能不相等；
比较两个struct时，只考虑其中的非blank的field是否相等；

slice、map和function value是不可以比较的(所以对应的类型值不能作为map的key)，但是它们可以和nil进行比较；同时也允许point、channel和interface value和nil进行比较；

# Logical operators

逻辑运算符对boolean值进行操作，运算符右边的操作数是条件执行的；

## Address operators

使用寻址运算符的操作数必须是可寻址的：
1. 变量；
2. 指针重定向后的值；
3. slice 索引值；
4. 可寻址struct的field；
5. 可寻址array的索引值；
6. 组合字面量

如果x是nil指针，访问*x将引起run-time panic;

## Receive operator

从channel接收数据的运算符表达式将block直到有数据可用；从nil channel接收数据将一直block；

从closed 的channel接收数据将立即返回，返回的结果为对应类型的zero value(当channel中所有的数据都被接收后)；
通过使用 多重赋值 表达式的第二个参数检查，channel是否closed；
	var x, ok = <-ch

receive operator是一个表达式，故可以用在各语句中；而向channel发送数据则是一条语句；

## Conversions

	Conversion = Type "(" Expression [ "," ] ")" .

转换属于PrimaryExpr，优先级比运算符 Operators 高；

如果类型以*或<-开始，或者类型以func开始且无返回值列表，则必须要用括号包围类型，以防止歧义：
	*Point(p)        // same as *(Point(p))
	(*Point)(p)      // p is converted to *Point
	<-chan int(c)    // same as <-(chan int(c))
	(<-chan int)(c)  // c is converted to <-chan int
	func()(x)        // function signature func() x
	(func())(x)      // x is converted to func()
	(func() int)(x)  // x is converted to func() int
	func() int(x)    // x is converted to func() int (unambiguous)

**常量**类型值x可以转换为类型T的场景：
1. x是类型T可代表(representable)的值； // int(1.2) 非法，因为整型不能代表浮点型常量；
2. x和T都是浮点类型；
3. x是整型，T是字符串类型：x是整型值(常量或者变量)时，可以转换为string类型，结果为x代表的Unicode code point的单字符串

	uint(iota)               // iota value of type uint
	float32(2.718281828)     // 2.718281828 of type float32
	complex128(1)            // 1.0 + 0.0i of type complex128
	float32(0.49999999)      // 0.5 of type float32
	float64(-1e-1000)        // 0.0 of type float64
	string('x')              // "x" of type string
	string(0x266c)           // "♬" of type string
	MyString("foo" + "bar")  // "foobar" of type MyString
	string([]byte{'a'})      // not a constant: []byte{'a'} is not a constant
	(*int)(nil)              // not a constant: nil is not a constant, *int is not a boolean, numeric, or string type
	int(1.2)                 // illegal: 1.2 cannot be represented as an int
	string(65.0)             // illegal: 65.0 is not an integer constant

**非常量**类型值x可以转换为类型T的场景：
1. x assignable to T；
2. x的类型和T有相同的underlying类型；
3. x的类型和T都是unnamed point，且它们的指针base类型有相同的underlying类型；
4. x和T都是整型或浮点型；
5. x和T都是虚数类型；
6. x是整型、byte或runce的slice，T是string类型；
7. x是string类型，T是byte或runce的slice类型；

整数和字符串间转换时，会改变原始值；其它情况的转换，只是修改了value的代表形式，value本身并不改变；

指针和整型是有区别的，unsafe package提供了相关的函数可以在两者之间转换。

### Conversions between numeric types

**非常量**类型数值间的转换：

1. 对于整形类型间的转换：
	1. 如果值是有符号整形，则符号位将扩展结果类型所需的宽度；
	2. 如果值是无符号整型，则用0扩展到结果类型所需的宽度；
		v := uint16(0x10F0) 
		int8(v)的结果是 1111 0000，最高位的1为符号位；
		uint32(int8(v)): 由于int8是有符号整形，故符号位1将扩展到32位，结果为0xFFFFFFF0;
2. 浮点数转为整数时，小数部分将被忽略；
3. 如果整形或浮点数转为浮点数，则结果精度取决于目标类型；

In all non-constant conversions involving floating-point or complex values, if the result type cannot represent the value 
the conversion succeeds but the result value is implementation-dependent.

## Constant expressions

常量表达式只包含常量操作数且在编译时执行；

Untyped boolean, numeric, and string constants may be used as operands wherever it is legal to use an operand of boolean, 
numeric, or string type, respectively. 
注意，这里对numeric没有区分是那种类型，所以在需要数值的地方就可以使用无类型的constant(int32、int64、float32、float64等)；

除了移位运算符，如果双目运算符两边均是无类型常量，则结果和运算符右边的无类型常量一致；

常量比较的结果是无类型的boolean常量；如果位运算的左侧是无类型常量，则结果是整形常量；其它无类型常量的运算结果是同类型的无类型常量；

	const a = 2 + 3.0          // a == 5.0   (untyped floating-point constant)
	const b = 15 / 4           // b == 3     (untyped integer constant)
	const c = 15 / 4.0         // c == 3.75  (untyped floating-point constant)
	const Θ float64 = 3/2      // Θ == 1.0   (type float64, 3/2 is integer division)
	const Π float64 = 3/2.     // Π == 1.5   (type float64, 3/2. is float division)
	const d = 1 << 3.0         // d == 8     (untyped integer constant)
	const e = 1.0 << 3         // e == 8     (untyped integer constant)
	const f = int32(1) << 33   // illegal    (constant 8589934592 overflows int32)
	const g = float64(2) >> 1  // illegal    (float64(2) is a typed floating-point constant)
	const h = "foo" > "bar"    // h == true  (untyped boolean constant)
	const j = true             // j == true  (untyped boolean constant)
	const k = 'w' + 1          // k == 'x'   (untyped rune constant)
	const l = "hi"             // l == "hi"  (untyped string constant)
	const m = string(k)        // m == "x"   (type string)
	const Σ = 1 - 0.707i       //            (untyped complex constant)
	const Δ = Σ + 2.0e-4       //            (untyped complex constant)
	const Φ = iota*1i - 1/1i   //            (untyped complex constant)

The divisor of a constant division or remainder operation must not be zero:

3.14 / 0.0   // illegal: division by zero

The values of typed constants must always be accurately representable as values of the constant type. 
The following constant expressions are illegal:

	uint(-1)     // -1 cannot be represented as a uint
	int(3.14)    // 3.14 cannot be represented as an int
	int64(Huge)  // 1267650600228229401496703205376 cannot be represented as an int64
	Four * 300   // operand 300 cannot be represented as an int8 (type of Four)
	Four * 100   // product 400 cannot be represented as an int8 (type of Four)

## Order of evaluation

在package级别，初始化依赖决定了各变量声明中各初始化表达式的执行顺序；
否则，当执行表达式、赋值、return语句等中包含的操作数时，函数调用、方法调用、通信操作的顺序是语法上自左向右；
注意：指的是function calls、method calls 和 communication operations 3种类型的操作是自左向右的。

	y[f()], ok = g(h(), i()+x[j()], <-c), k()

上面的函数调用、通信操作的顺序是： f(), h(), i(), j(), <-c, g(), and k().
但是对于x的索引操作和y的索引操作的执行顺序是未定义的；

	a := 1
	f := func() int { a++; return a }

	x := []int{a, f()}            // x may be [1, 2] or [2, 2]: evaluation order between a and f() is not specified
	m := map[int]int{a: 1, a: 2}  // m may be {2: 1} or {2: 2}: evaluation order between the two map assignments is not specified
	n := map[int]int{a: f()}      // n may be {2: 3} or {3: 3}: evaluation order between the key and the value is not specified

在 package level，初始化依赖决定各初始化表达式的执行顺序(而不是默认的自左向右规则)；

	var a, b, c = f() + v(), g(), sqr(u()) + v()

	func f() int        { return c }
	func g() int        { return a }
	func sqr(x int) int { return x*x }

	// functions u and v are independent of all other variables and functions

函数调用的执行顺序是u()、sqr()、v()、f()、v()和g()；

# Statements

语句控制执行；

	Statement =
		Declaration | LabeledStmt | SimpleStmt |
		GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
		FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
		DeferStmt .

	SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .

## Terminating statements

终止语句指的是终止后续语句的执行，包括：

1. A "return" or "goto" statement.
2. A call to the built-in function panic.
3. A block in which the statement list ends in a terminating statement.
4. An "if" statement in which:
	+ the "else" branch is present, and
	+ both branches are terminating statements.
5. A "for" statement in which: ；
	+ there are no "break" statements referring to the "for" statement, and
	+ the loop condition is absent. // for死循环终止后续语句的执行
6. A "switch" statement in which:
	+ there are no "break" statements referring to the "switch" statement,
	+ there is a default case, and
	+ the statement lists in each case, including the default, end in a terminating statement, or a possibly labeled "fallthrough" statement.  // 各个case(包括default)必须都终止，该switch才能终止后续语句的执行；
7. A "select" statement in which:
	+ there are no "break" statements referring to the "select" statement, and
	+ the statement lists in each case, including the default if present, end in a terminating statement. // 各个case(包括default)必须都终止，该select才能终止后续语句的执行；
8. A labeled statement labeling a terminating statement.

All other statements are not terminating.

A statement list ends in a terminating statement if the list is not empty and its final non-empty statement is terminating.

## Empty statements

The empty statement does nothing.

	EmptyStmt = .

## Labeled statements

A labeled statement may be the target of a goto, break or continue statement.

	LabeledStmt = Label ":" Statement .
	Label       = identifier .

	Error: log.Panic("error encountered")

经常用于 for 关键字的前面，用于跳出多重循环；

## Expression statements

ExpressionStmt = Expression .

除了下列的内置函数外，函数、方法和接收运算符可以出现在语句上下文中(单行)；

append cap complex imag len make new real
unsafe.Alignof unsafe.Offsetof unsafe.Sizeof

如果直接调用上面的内置函数，会提示：./test.go:4: len("1234") evaluated but not used

	h(x+y)
	f.Close()
	<-ch
	(<-ch)
	len("foo")  // illegal if len is the built-in function

## Send statements

	SendStmt = Channel "<-" Expression .
	Channel  = Expression .

channel 和 value 表达式在通信前执行；
向关闭的 channel 发数据会引起 runtime panic;
向 nil channel 发数据会一直被block，编译器会检查到这种状态，提示panic：
	zhangjun@localhost:~% cat test.go
	package main

	func main() {
		var c chan int
		c <- 1
	}
	zhangjun@localhost:~% go run test.go
	fatal error: all goroutines are asleep - deadlock!

	goroutine 1 [chan send (nil chan)]:
	main.main()
		/Users/zhangjun/test.go:5 +0x49
	exit status 2

## IncDec statements
	IncDecStmt = Expression ( "++" | "--" ) .

Expression 的结果必须是可寻址的，或者是 map index expression；
注意：自增自减是语句而非表达式；

## Assignments

	Assignment = ExpressionList assign_op ExpressionList .

	assign_op = [ add_op | mul_op ] "=" .

等号左边的操作数必须是可寻址的、map index表达式或者blank标识符；

	x = 1
	*p = f()
	a[i] = 23
	(k) = <-ch  // same as: k = <-ch

An assignment operation x op= y where op is a binary arithmetic operation is equivalent to x = x op (y) but evaluates x only once. 

赋值运算符按两阶段执行：
1. 等号左边的索引表达式、指针重定向，等号右边的表达式按顺序执行；
2. 将执行结果从左向右赋值；

	a, b = b, a  // exchange a and b

	x := []int{1, 2, 3}
	i := 0
	i, x[i] = 1, 2  // set i = 1, x[0] = 2

	i = 0
	x[i], i = 2, 1  // set x[0] = 2, i = 1

	x[0], x[0] = 1, 2  // set x[0] = 1, then x[0] = 2 (so x[0] == 2 at end)

	x[1], x[3] = 4, 5  // set x[1] = 4, then panic setting x[3] = 5.

	type Point struct { x, y int }
	var p *Point
	x[2], p.x = 6, 7  // set x[2] = 6, then panic setting p.x = 7

	i = 2
	x = []int{3, 5, 7}
	for i, x[i] = range x {  // set i, x[2] = 0, x[0]
		break
	}
	// after this loop, i == 0 and x == []int{3, 5, 3}

## If statements
	IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .

	Expression后面是Blcok，故必须使用大括号扩住语句；

## Switch statements
	SwitchStmt = ExprSwitchStmt | TypeSwitchStmt .

### Expression switches

	ExprSwitchStmt = "switch" [ SimpleStmt ";" ] [ Expression ] "{" { ExprCaseClause } "}" .
	ExprCaseClause = ExprSwitchCase ":" StatementList .
	ExprSwitchCase = "case" ExpressionList | "default" .

如果 switch 表达式结果是无类型常量，则首先被转换为对应的缺省类型； nil 不能作为 switch expression;
如果 case 表达式是无类型的x，则先被转换为 switch expression 的类型t，x 和 t 必须是可比较的。
各 case 的语句列表的最后一条可以是 fallthrough 语句，表示控制流转移到下一条 case 子句，否则执行完 case 子句后，将跳转到 switch 语句的结束为止；

	switch tag {
	default: s3()
	case 0, 1, 2, 3: s1()
	case 4, 5, 6, 7: s2()
	}

	switch x := f(); {  // missing switch expression means "true"
	case x < 0: return -x
	default: return x
	}

	switch {
	case x < y: f1()
	case x < z: f2()
	case x == 4: f3()
	}

### Type switches

type switch 比较的是类型而非值，它使用类型断言的形式，但是类型是关键字 type 而非实际类型：

	switch x.(type) {
	// cases
	}

x 必须是接口类型，case 列出的类型必须是各不相同的；

	TypeSwitchStmt  = "switch" [ SimpleStmt ";" ] TypeSwitchGuard "{" { TypeCaseClause } "}" .
	TypeSwitchGuard = [ identifier ":=" ] PrimaryExpr "." "(" "type" ")" .
	TypeCaseClause  = TypeSwitchCase ":" StatementList .
	TypeSwitchCase  = "case" TypeList | "default" .
	TypeList        = Type { "," Type } .

TypeSwitchGuard 可能包含一个短变量声明，相当于在每个 case 子句前面声明了一个变量；如果 case 子句的 TypeList 只包含一个类型，则该变量即为相应的类型，
否则为接口类型；

示例: 
	switch i := x.(type) {
	case nil:
		printString("x is nil")                // type of i is type of x (interface{})
	case int:
		printInt(i)                            // type of i is int
	case float64:
		printFloat64(i)                        // type of i is float64
	case func(int) float64:
		printFunction(i)                       // type of i is func(int) float64
	case bool, string:
		printString("type is bool or string")  // type of i is type of x (interface{})
	default:
		printString("don't know the type")     // type of i is type of x (interface{})
	}

type switch 不支持 fallthrough 语句；

## For statements

	ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
	Condition = Expression .

for a < b {
	a *= 2
}

如果 Condition 为空，则相当于条件一直为 true；

	ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
	InitStmt = SimpleStmt .
	PostStmt = SimpleStmt .

三个子句可以省略，但是后面的分号不能省略；

for i := 0; i < 10; i++ {
	f(i)
}

for cond { S() }    is the same as    for ; cond ; { S() }
for      { S() }    is the same as    for true     { S() }

	RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .

注意： range左侧的List是可选的，可以为空；

range 右边的表达式可以为 array、array的指针、slice、string、map 和 接收数据的 channel；表达式左边的操作数必须是可寻址、map index 表达式；

range 表达式只在 loop 执行前执行一次；如果表达式是 array、array指针，且 range 左边只有一个变量，则只会执行 range 表达式的 length 函数；

range 左边的表达式列表或标识符列表可以为空：
	// empty a channel
	for range ch {}

for range可以迭代nil的map，立即返回；

## Go statements

	GoStmt = "go" Expression .

表达式必须是函数或者方法调用。

## Select statements

	SelectStmt = "select" "{" { CommClause } "}" .
	CommClause = CommCase ":" StatementList .
	CommCase   = "case" ( SendStmt | RecvStmt ) | "default" .
	RecvStmt   = [ ExpressionList "=" | IdentifierList ":=" ] RecvExpr . // 接收语句可以定义新的变量；
	RecvExpr   = Expression .

select 语句的执行步骤：
1. 按照源文件顺序执行 channel 操作数(包括发送和接收)，且只执行一次；结果是 可以发送或接收的 channel 和发送的数据；在此过程中可能产生副作用，但与选择哪个 case 执行无关；等号左边的
表达式在此阶段并未执行；
2. 如果有一个或多个通信可以处理，则随机的选择一个。否则如果有default，则选择default；如果没有default，则select阻塞，直到有通信可以进行；
3. 执行相应的通信操作(default 子句除外)；
4. 如果选择的子句包含短变量的从 channel 接收数据，则执行表达式左边的表达式；
5. 执行选择的 case 包含的语句列表；

select 只包含 nil channel 且无 default 子句会一直被 block ；

	for {  // send random sequence of bits to c
		select {
		case c <- 0:  // note: no statement, no fallthrough, no folding of cases
		case c <- 1:
		}
	}

	select {}  // block forever

## Return statements

	ReturnStmt = "return" [ ExpressionList ] .

## Break statements

	BreakStmt = "break" [ Label ] .

## Continue statements

	ContinueStmt = "continue" [ Label ] .

## Goto statements

	GotoStmt = "goto" Label .

## Fallthrough statements

	FallthroughStmt = "fallthrough" .

## Defer statements

	DeferStmt = "defer" Expression .

# Built-in functions
内置函数没有对应的标准 Go 类型，因此只能出现在函数调用表达式中，不能作为变量传递；

## Close
向已关闭的 channel 发送数据会引起 panic；关闭 nil channel 也会引起 panic；

## Length and capacity
nil slice、map、channel 的长度是 0， nil slice、channel 的容量也是 0；

## Allocation
new(T)

## Making slices, maps and channels
make 接收一个类型是 slice、map 或 channel 的类型 T，返回一个类型为 T(非 *T) 的值；

## Appending to and copying slices
append 和 copy 适用于 slice 类型，

append 返回的 slice 有可能不是传入的第一个参数 slice；

	s0 := []int{0, 0}
	s1 := append(s0, 2)                // append a single element     s1 == []int{0, 0, 2}
	s2 := append(s1, 3, 5, 7)          // append multiple elements    s2 == []int{0, 0, 2, 3, 5, 7}
	s3 := append(s2, s0...)            // append a slice              s3 == []int{0, 0, 2, 3, 5, 7, 0, 0}
	s4 := append(s3[3:6], s3[2:]...)   // append overlapping slice    s4 == []int{3, 5, 7, 2, 3, 5, 7, 0, 0}

	var t []interface{} // t 未被初始化，但是可以用于 append；
	t = append(t, 42, 3.1415, "foo")   //                             t == []interface{}{42, 3.1415, "foo"}

	var b []byte
	b = append(b, "bar"...)            // append string contents      b == []byte{'b', 'a', 'r' }

copy(dst, src []T) int
copy(dst []byte, src string) int

copy 将 src 中的元素 copy 到 dst 开头的位置，copy 的元素数目是 dst、src 最小长度；

## Deletion of map elements
delete(m, k)  // remove element m[k] from map m

从 map 中移除 key 对应的元素；如果 m 为 nil 或者不包含 k，则啥都不干；

## Manipulating complex numbers
	complex(realPart, imaginaryPart floatT) complexT
	real(complexT) floatT
	imag(complexT) floatT

## Handling panics
	func panic(interface{})
	func recover() interface{}

func protect(g func()) {
	defer func() {
		log.Println("done")  // Println executes normally even if there is a panic
		if x := recover(); x != nil {
			log.Printf("run time panic: %v", x)
		}
	}()
	log.Println("start")
	g()
}

recover 函数返回 nil 的情况：
1. panic's argument was nil;
2. the goroutine is not panicking;
3. recover was not called directly by a deferred function.

## Bootstrapping
当前语言实现的几个内置函数：

print      prints all arguments; formatting of arguments is implementation-specific
println    like print but prints spaces between arguments and a newline at the end

# Packages 
Go 程序是由 package 链接而成的；pacakge 是由一个或多个源文件组成；

## Source file organization
	SourceFile       = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .

## Package clause
PackageClause  = "package" PackageName .
PackageName    = identifier .

## Import declarations
ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
ImportSpec       = [ "." | PackageName ] ImportPath .
ImportPath       = string_lit .

Import declaration          Local name of Sin

import   "lib/math"         math.Sin
import m "lib/math"         m.Sin
import . "lib/math"         Sin

To import a package solely for its side-effects (initialization), use the blank identifier as explicit package name:

import _ "lib/math"

# Program initialization and execution
## The zero value
Each element of such a variable or value is set to the zero value for its type: 
false for booleans, 0 for integers, 0.0 for floats, "" for strings, and nil for pointers, functions, interfaces, 
slices, channels, and maps. This initialization is done recursively, so for instance each element of an array of 
structs will have its fields zeroed if no value is specified.

## Package initialization
Within a package, package-level variables are initialized in declaration order but after any of the variables they depend on.

## Program execution
A complete program is created by linking a single, unimported package called the main package with all the packages it imports,
transitively.

# Errors
type error interface {
	Error() string
}

## Run-time panics
package runtime

type Error interface {
	error
	// and perhaps other methods
}

# System considerations
## Package unsafe
	package unsafe

	type ArbitraryType int  // shorthand for an arbitrary Go type; it is not a real type
	type Pointer *ArbitraryType

	func Alignof(variable ArbitraryType) uintptr
	func Offsetof(selector ArbitraryType) uintptr
	func Sizeof(variable ArbitraryType) uintptr

A Pointer is a pointer type but a Pointer value may not be dereferenced. Any pointer or value of underlying type uintptr 
can be converted to a Pointer type and vice versa. The effect of converting between Pointer and uintptr is 
implementation-defined.

	var f float64
	bits = *(*uint64)(unsafe.Pointer(&f))

	type ptr unsafe.Pointer
	bits = *(*uint64)(ptr(&f))

	var p ptr = nil

## Size and alignment guarantees
	type                                 size in bytes

	byte, uint8, int8                     1
	uint16, int16                         2
	uint32, int32, float32                4
	uint64, int64, float64, complex64     8
	complex128                           16