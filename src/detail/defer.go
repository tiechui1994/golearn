package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// defer 执行顺序:
// 原则: 从"抛出异常位置或者函数执行结束的位置"向前, 离其位置越近, defer越先执行. 向后的defer不执行

// 打印顺序: 打印后 打印中 打印前
func DeferExecute() {
	defer func() {
		fmt.Println("打印前")
	}()
	defer func() {
		fmt.Println("打印中")
	}()
	panic("触发异常")

	defer func() {
		fmt.Println("打印后")
	}()
}

// defer 函数参数初始化:
// 当程序运行到defer函数时, 不会执行函数体, 但会将defer函数中的参数进行进行实例化.
func DeferInit() {
	calc := func(index string, a, b int) int {
		ret := a + b
		fmt.Println(index, a, b, ret)
		return ret
	}

	var (
		a = 1
		b = 2
	)
	defer calc("1", a, calc("10", a, b)) // 参数: a=1, calc(a=1, b=2)
	a = 0
	defer calc("2", a, calc("20", a, b)) // 参数: a=0, calc(a=0, b=2)
	b = 1
}

/*
defer 函数体:
	在运行defer函数体的时机, 会顺序执行函数体的内容
*/
func DeferParam() {
	calc := func(index string, a, b int) int {
		ret := a + b
		fmt.Println(index, a, b, ret)
		return ret
	}

	var (
		a = 1
		b = 2
	)
	defer func() {
		calc("1", a, calc("10", a, b)) // 参数: a=0, calc(a=0, b=1)
	}()
	a = 0
	defer func() {
		calc("2", a, calc("20", a, b)) // 参数: a=0, calc(a=0, b=1)
	}()
	b = 1
}

/*
defer函数 与 函数变量的作用域:
	1. defer需要在函数结束前执行.(具体: 函数体执行 -> defer函数执行 -> return)
    2. 函数返回值名字会在函数起始处为赋予对应类型的零值并且作用域为整个函数(包括内部函数).
	3. 函数内的变量的作用域是当前函数, 注意:内部函数可以引用变量的值, 但是无法修改其值(即使是使用指针进行操作, 或者通过参数传入其指针)
	4. 当一个函数有返回值变量时, 其最终返回值变量就确定了. return realValue 和 return varValue 都是对返回值变量的隐式初始化.

函数返回值:
	1. 在函数有多个返回值时, 只要有一个返回值有指定命名, 其他的也必须有命名.
	2. 如果返回值有有多个, 则返回值必须加上括号
	3. 如果只有一个返回值并且有命名, 则必须加上括号

nil:
	nil可以用作 interface、function、pointer、map、slice 和 channel 的"空值"
*/
type test struct {
	value int
}

func (t *test) Close() {
	t.value = -1
	fmt.Printf("this is close\n")
}

func DeferAndScope() {
	f1 := func(i int) (t int) {
		t = i
		defer func() {
			t += 3
		}()
		return t
	}

	f2 := func(i int) int {
		t := i
		defer func() {
			t += 3
		}()
		return t
	}

	f3 := func(i int) (t int) {
		t = 10
		defer func() {
			fmt.Println("-----", t)
			t += i
		}()
		return 2 // 隐藏赋值操作
	}

	f4 := func() *test {
		t := &test{value: 100}
		defer func() {
			t.Close()
		}()
		return t
	}

	f5 := func() (t test) {
		t = test{value: 100}
		defer func() {
			t.Close()
		}()
		return t
	}

	fmt.Println(f1(1))
	fmt.Println(f2(1))
	fmt.Println(f3(1))
	fmt.Println("==============================")
	fmt.Println(f4())
	fmt.Println(f5())
}

/*
golang 作用域:
	全局作用域: 所有的关键字和内置类型, 内置函数都拥有全局作用域
	package作用域: 一个包中声明的变量, 常量, 函数, 类型等都拥有包作用域, 在同一个包中可以任意访问
	文件作用域: 一个文件中通过import导入的包名, 只在当前文件有效
	函数作用域以及for, if, {}自定义作用域: 函数的参数, 命名函数返回值, 返回值都拥有函数作用域, for,
	if, {}内部的变量只拥有最小作用域.
*/
func Scope() {
	var ErrDidNotWork = errors.New("did not work")

	tryTheThing := func() (string, error) {
		return "", ErrDidNotWork
	}

	DoTheThing := func(reallyDoIt bool) (err error) {
		if reallyDoIt {
			result, err := tryTheThing()
			if err != nil || result != "it worked" {
				err = ErrDidNotWork
			}
		}
		return err
	}
	fmt.Println(DoTheThing(true))
	fmt.Println(DoTheThing(false))
}

/*
可以使用 == 进行比较的类型:
	int, string, 指针, channel, 简单结构体(结构体的元素只包含int, string, 指针, channel), 且是相同类型(元素名称,元素类型,元素顺序严格一致)

不可以使用 == 进行比较:
	slice, 数组, 复杂结构体, map
*/
func EqualStruct() {
	// 指针
	a1 := &[]int{1, 2}
	b1 := &[]int{1, 2}
	if a1 == b1 {
		println("==")
	}

	// 指针
	a2 := new(map[string]string)
	b2 := new(map[string]string)
	if a2 == b2 {
		println("===")
	}
}

// make创建切片, 会为切片分配好n个切片类型的空间, 并且初始值为零值
// 接下来, 使用数组赋值的操作会修改原来的零值; 使用append操作会扩展原来的切片, 并添加元素
func Slice() {
	s := make([]int, 5)
	s = append(s, 1, 2, 3)
	fmt.Println(s)
}

/*
golang当中的map不是线程安全的, 在并发读写map的时候会出现"concurrent map read and map write" 错误.
但是golang当中的array是线程安全的.
*/
type Map struct {
	mp map[string]int
	//sync.Mutex
}

func (ua *Map) Add(name string, age int) {
	//ua.Lock()
	//defer ua.Unlock()
	ua.mp[name] = age
}

func (ua *Map) Get(name string) int {
	//ua.Lock()
	//defer ua.Unlock()
	if age, ok := ua.mp[name]; ok {
		return age
	}
	return -1
}

func ConcurrentMap() {
	u := Map{mp: make(map[string]int)}
	channel := make(chan string, 16)
	getString := func(length int) string {
		bytes := make([]byte, length)
		rand.Read(bytes)
		return hex.EncodeToString(bytes)
	}

	for i := 0; i < 5; i++ {
		go func(channel chan string) {
			for {
				str := getString(5)
				channel <- str
				fmt.Println("Add: ", str)
				u.Add(str, 21)
			}
		}(channel)
	}

	go func(channel chan string) {
		for {
			str := <-channel
			age := u.Get(str)
			fmt.Println("Get: ", str, age)
		}
	}(channel)

	time.Sleep(10 * time.Second)
}

/*
vars.(type) 使用的前提是:
	1. vars必须是显示声明的接口类型.
	2. 该表达式只能使用在switch...case结构当中

vars.(int/string/struct{}/interface{}) 使用的前提是:
	1. vars必须是显示声明的接口类型.
*/
func UseType() {
	var types interface{} = 10
	switch types.(type) {
	case int:
		print("int")
	default:
		print("un")
	}

	var intvar interface{} = 100
	print(intvar.(int))
}

/*
常量初始化: 参考例子
	1. 局部的常量可以覆盖全局常量
	2. 全局常量可以使用 _ 跳过不想设置的值(iota), 但是局部常量办不到

内存地址:
	变量(短变量和声明变量)具有内存地址
	常量没有内存地址
*/
func ConstValue() {
	const (
		a = iota // 0
		b        // 1
		c = 1    // 1
		d        // 1
		e = iota // 4
		f        // 5
		g        // 8
		h = "a"  // "a"
		i        // "a"
		l = iota // 11
	)
	println(a, b, c, d, e, f, g, h, i, l)
}

/*
var关键字的使用:
	1. 声明式变量 等价于 零值初始化式变量
	2. nil 是 interface、function、pointer、map、slice 和 channel 的"零值"
*/
func VarValue() {
	var (
		a struct {
			// 匿名类型的声明式变量
			Id int
		}
		b int     // int变量
		c *string // 指针变量
		d = struct {
			// 初始化式变量
			Id int
		}{}
	)

	fmt.Printf("%+v, %+v, %+v, %+v\n", a, b, c, d)
}
