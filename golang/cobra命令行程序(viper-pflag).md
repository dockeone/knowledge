# 介绍

Cobra是一个可以创建强大命令行接口程序的库，同时也可以生成命令行接口程序框架的程序，它支持：

1. 子命令：如 app server、app fetch等；
2. 完整地支持POSIX、GNU风格的命令行参数(包括短和长版本，基于github.com/ogier/pflag实现)；
3. 支持嵌套的子命令，嵌套级别无限制；
4. 支持全局、本地等层叠(cascading) flags，也就是说本地没有设置，则一层层向上搜索，直到找到或使用缺省值；
5. cobra create appname和 cobra add cmdname 可以用来生成命令行程序的框架；
6. 智能提示( app server ... did you mean app server?)
7. 自动生成命令和flags的帮助文档；
8. 指定生成detail help如 app help [command]
9. 自动生成autocomplate和man pages；
10. 自动上传帮助选项 -h、--help等；
11. 支持command alias；
12. 支持自定义help、usage等；
13. 结合viper，可以支持12-factor apps;
14. flags可以放在command前或者后(只要不引起歧义即可)，长和短flags可以同时使用，支持命令简写(唯一即可)；

#概念

Cobra is built on a structure of commands, arguments & flags.
Commands represent actions, Args are things and Flags are modifiers for those actions.

The pattern to follow is APPNAME VERB NOUN --ADJECTIVE. or APPNAME COMMAND ARG --FLAG

A Cobra command can define flags that persist through to children commands and flags that are only available to that command.

Cobra works by creating a set of commands and then organizing them into a tree. The tree defines the structure of the application.

Once each command is defined with its corresponding flags, then the tree is assigned to the commander which is finally executed.

#POSIX和GNU风格的flags

go的flag使用的是 plan9风格的命令行选项: 短横杠开始的多个字母如-enable，flag也支持两个短横杠的选项(boolean类型的选项除外)；

http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html
POSIX建议的是单字母的短命令行格式，支持合并：
+ Arguments are options if they begin with a hyphen delimiter (‘-’).
+ Multiple options may follow a hyphen delimiter in a single token if the options do not take arguments. Thus, ‘-abc’ is equivalent to ‘-a -b -c’.
+ Option names are single alphanumeric characters (as for isalnum; see Classification of Characters).
+ Certain options require an argument. For example, the ‘-o’ command of the ld command requires an argument—an output file name.
+ An option and its argument may or may not appear as separate tokens. (In other words, the whitespace separating them is optional.) 
    Thus, ‘-o foo’ and ‘-ofoo’ are equivalent.p
+ Options typically precede other non-option arguments.
+ The argument ‘--’ terminates all options; any following arguments are treated as non-option arguments, even if they begin with a hyphen.
+ A token consisting of a single hyphen character is interpreted as an ordinary non-option argument. By convention, it is used to specify input from or output to the standard input and output streams.
+ Options may be supplied in any order, or appear multiple times. The interpretation is left up to the particular application program.

GNU在POSIX风格的基础上加了long options惯例:

1. 参数名一般是2-3个单词，使用短横杠划分各单词；
2. 用户可以简写参数名称，只要该简写是唯一的；
3. 如果需要给long options指定argument, 使用 "--name=value"的形式；


#cobra使用的ogier/pflag的命令行flag语法：

--flag    // boolean flags, or flags with no option default values
--flag x  // only on flags without a default value
--flag=x

Unlike the flag package, a single dash before an option means something different than a double dash. Single dashes signify a series of shorthand 
letters for flags. All but the last shorthand letter must be boolean flags or a flag with a default value

// boolean or flags where the 'no option default value' is set
-f
-f=true
-abc
but
-b true is INVALID

// non-boolean and flags without a 'no option default value'
-n 1234
-n=1234
-n1234

// mixed
-abcs "hello"
-absd="hello"
-abcs1234

Flag parsing stops after the terminator "--". Unlike the flag package, flags can be interspersed with arguments anywhere on the command line 
before this terminator.

Integer flags accept 1234, 0664, 0x1234 and may be negative. Boolean flags (in their long form) accept 1, 0, t, f, true, false, TRUE, FALSE, 
True, False. Duration flags accept any input valid for time.ParseDuration.

#API示例：

1. 定义root命令

    var RootCmd = &cobra.Command{
        Use:   "netbench",
        Short: `网络质量监控系统`,
        Long: `高性能网络质量监控系统
        `
        Run:   Root,
    }

// 注意如果要执行./netbench --opt选项，则必须定义RootCmd的Run函数，然后在Run函数里解析选项；
// 如下面函数处理--version选项: ./netbench --version
func Root(cmd *cobra.Command, args []string) {
	if viper.GetBool("version") {
		fmt.Printf("%s\n", VERSION)
	}
}

func init() {
    RootCmd.AddCommand(master.MasterCmd) // 添加子命令
}

定义子命令：
var MasterCmd = &cobra.Command{
        Use:   "masterUse",  //子命令名称
        Short: "Short: master节点", //子命令的简短描述
        Long: `netbench longlongllllllllllllllll //打印该命令帮助时显示的详细信息
        description`,
        Run:   master,
}

$ ./netbench  -h
网络质量监控系统

Usage:
  netbench [command]

Available Commands:
  masterUse   Short: master节点   //分别对应子命令的Use和Short
  slave       slave节点
  pcap        pcap节点
  alert       alert节点

$ ./netbench masterUse -h
netbench longlongllllllllllllllll   //对应子命令的Long参数
description

Usage:
  netbench masterUse [flags]


2. 初始化root命令的参数

cmd.PersistentFlags()和cmd.Flags()返回的是*pflag.FlagSet类型(github.com/ogier/pflag包), 该类型兼容go标准库中的flag package，另外定义了以P结尾的方法，用来同时支持长和短命令
行选项：

func init() {
    //cobra.OnInitialize注册在调用每个子命令的preRun阶段自动调用的初始化函数；
    //func (c *Command) preRun() {
	//  for _, x := range initializers {
	//	  x()
	//  }
    //}
    cobra.OnInitialize(initConfig)

    //PersistentFlags会被root命令和子命令继承；
    RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)") //注意缺省值里面不能使用环境变量如$HOME;
    RootCmd.PersistentFlags().StringVarP(&projectBase, "projectbase", "b", "", "base project directory eg. github.com/spf13/")
    RootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
    RootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project (can provide `licensetext` in config)")
    RootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
    
    //Local Flags 只对RootCmd有效；
    RootCmd.Flags().StringVarP(&Source, "source", "s", "", "Source directory to read from")

    //将viper的name和上面定义的Flag绑定；
    viper.BindPFlag("author", RootCmd.PersistentFlags().Lookup("author"))
    viper.BindPFlag("projectbase", RootCmd.PersistentFlags().Lookup("projectbase"))
    viper.BindPFlag("useViper", RootCmd.PersistentFlags().Lookup("viper"))

    //如果上面的Lookup找不到对应的值，则viper使用下面定义的缺省值；
    viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
    viper.SetDefault("license", "apache")
}

//初始化函数，在调用每个子命令的preRun阶段执行；
func initConfig() {
	cfgFile := viper.GetString("cfgFile")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// SetConfigName会将configFile设置为空，故两者不能同时调用
		viper.SetConfigName(".netbench") // 不需要加后缀名
	}
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

3. 定义子命令

package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    //将子命令注册到父命令
    RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number of Hugo",
    Long:  `All software has versions. This is Hugo's`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
    },
}

4. 主函数，Execute root命令；

package main

import "{pathToYourApp}/cmd"

func main() {
    if err := cmd.RootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }
}

#其它特性

1. 支持PreRun和PostRun Hooks：
+ PersistentPreRun
+ PreRun
+ Run
+ PostRun
+ PersistentPostRun

PersistentPreRun和PersistentPostRun会被子命令继承(如果子命令没有定义对应的函数)；

2. cmd.Execute()默认读取的是os.Args，可以使用它的SetArgs方法设置它的命令行参数(除了命令本身)：

rootCmd.AddCommand(subCmd)

rootCmd.SetArgs([]string{""})
_ = rootCmd.Execute()
fmt.Print("\n")
rootCmd.SetArgs([]string{"sub", "arg1", "arg2"})
_ = rootCmd.Execute()

3. 如果解析命令行参数出现了错误(如给命令或子命令传入了错误的选项、参数)，可以给command定义Error处理Hooks函数：

+ PersistentPreRunE
+ PreRunE
+ RunE
+ PostRunE
+ PersistentPostRunE

如：
func main() {
    var rootCmd = &cobra.Command{
        Use:   "hugo",
        Short: "Hugo is a very fast static site generator",
        Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Do Stuff Here
            return errors.New("some random error")
        },
    }

    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

#viper支持12-factor类型的app获取参数：

1. setting defaults
2. reading from JSON, TOML, YAML, HCL, and Java properties config files，无需关心文件格式；
3. live watching and re-reading of config files (optional)，监视配置文件变化，动态重载配置参数；
4. reading from environment variables，从环境变量读取参数；
5. reading from remote config systems (etcd or Consul), and watching changes
6. reading from command line flags，和cobra结合，读取命令行参数；
7. reading from buffer
8. setting explicit values

各参数源的优先级如下(如果设置相应的源，且设置了相应的参数值)：

1. explicit call to Set
2. flag
3. env
4. config
5. key/value store
6. default

设置viper参数值

1. 设置缺省值：
viper.SetDefault("LayoutDir", "layouts")
viper.SetDefault("Taxonomies", map[string]string{"tag": "tags", "category": "categories"})

2. 从配置文件中读取：

viper.SetConfigName("config") // name of config file (without extension)
viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
viper.AddConfigPath(".")               // optionally look for config in the working directory

err := viper.ReadInConfig() // Find and read the config file
if err != nil { // Handle errors reading the config file
    panic(fmt.Errorf("Fatal error config file: %s \n", err))
}

3. 监视并从配置文件中读取(确保如步骤2所示，添加了所有的配置文件路径)：

viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    fmt.Println("Config file changed:", e.Name)
})

4. 从io.Reader中读取：

viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

// any approach to require this configuration into your program.
var yamlExample = []byte(`
Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
age: 35
eyes : brown
beard: true
`)

viper.ReadConfig(bytes.NewBuffer(yamlExample))

viper.Get("name") // this would be "steve"

5. 设置配置参数，结果优先级最高：

viper.Set("Verbose", true)
viper.Set("LogFile", LogFile)

6. 设置别名

viper.RegisterAlias("loud", "Verbose")
viper.Set("verbose", true) // same result as next line
viper.Set("loud", true)   // same result as prior line
viper.GetBool("loud") // true
viper.GetBool("verbose") // true

7. 使用环境变量，支持4种类型的环境变量：

+ AutomaticEnv()
+ BindEnv(string...) : error
+ SetEnvPrefix(string)
+ SetEnvReplacer(string...) *strings.Replacer

viper对环境变量是大小写敏感的；
通过SetEnvPrefix()函数，可以设置ENV的前缀，AutomaticEnv和BindEnv会使用这个前缀；

BindEnv：传入一个或两个参数，第一个参数为key name，第二个参数为ENV name，如果只传入一个参数，则为key name，ENV name为 KEY，即key name的大写形式；

如果传入了ENV name，则viper不会自动加Prefix；viper 获取环境变量值时，每次都重新获取，而不是调用BindEnv时的值；

AutomaticEnv一般和SetEnvPrefix结合使用，调用后，当viper.Get获取key值时，viper会检查加了Prefix的全大写的KEY是否存在；

8. 使用Flags：

viper只是cobra库使用的pflags；和BindEnv一样，该value不是在调用Bind方法时设置，而是当调用viper.Get时获取；

9. 从viper获取key对应的value：

+ Get(key string) : interface{}
+ GetBool(key string) : bool
+ GetFloat64(key string) : float64
+ GetInt(key string) : int
+ GetString(key string) : string
+ GetStringMap(key string) : map[string]interface{}
+ GetStringMapString(key string) : map[string]string
+ GetStringSlice(key string) : []string
+ GetTime(key string) : time.Time
+ GetDuration(key string) : time.Duration
+ IsSet(key string) : bool


常见错误：
1. 在top作用域如var中使用viper.GetString()如：
    var sqliteDbPath  = viper.GetString("master.sqlite3.db")
这种使用方法是不正确的，因为一般情况下是在init函数中BindFlag如：
func init (){
    viper.BindPFlag("master.sqlite3.db", AlertCmd.Flags().Lookup("sqliteDbPath"))
}
而init是在var变量初始化执行后才执行的，导致上面的sqliteDbPath结果为空；

2. 多个位置的init函数重复对同一个key设置BindFlag，如：
master.go:
    viper.BindPFlag("masterURL", SlaveCmd.Flags().Lookup("masterURL"))
slave.go:
    viper.BindPFlag("masterURL", SlaveCmd.Flags().Lookup("masterURL"))
这样的话，会以最后一次设置的位置；

3. 同时调用viper.SetConfigFile(cfgFile)和viper.SetConfigName(".netbench")
    viper.SetConfigName()函数将configFile设置为空，所以如果同时调用这两个函数，SetConfigFile()设置的配置文件会失效：
   正确的调用方式是：
   // 初始化viper
    func initConfig() {
        cfgFile := viper.GetString("cfgFile")
        if cfgFile != "" {
            viper.SetConfigFile(cfgFile)
        } else {
            // SetConfigName会将configFile设置为空，故两者不能同时调用
            viper.SetConfigName(".netbench") // 不需要加后缀名
        }
        viper.AddConfigPath("$HOME")
        viper.AddConfigPath(".")
        viper.AutomaticEnv()

        if err := viper.ReadInConfig(); err == nil {
            log.Println("Using config file:", viper.ConfigFileUsed())
        }
    }

4. viper.GetStringSlice()的使用方式如下：
    func init() {
        SlaveCmd.Flags().StringSlice("isps", []string{"all"}, "监控的ISP类型")
    }

    func Slave(stop, done chan struct{}) {
        if strings.Contains(viper.GetStringSlice("slave.isps")[0], ",") {
            SlaveOpts.isps = strings.Split(viper.GetStringSlice("slave.isps")[0], ",")
        } else {
            SlaveOpts.isps = viper.GetStringSlice("slave.isps")
        }
    }
 命令行参数如下：
    ./netbench slave --isps "电信 联通 移动" // pflags使用空格分割选项值；
  和：
    ./netbench slave --isps 电信 --isps 移动
  是等效的。viper.GetStringSlice("isps")返回的是 []string{"电信,移动,联通"} 故需要Split;
  如果在配置文件(如YAML)中指定了isps列表，则viper.GetStringSlice("isps")返回的是正常的 []string{"电信","移动","联通"}
  所以上面的代码中会分情况进行处理；https://github.com/spf13/viper/issues/112