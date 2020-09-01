## reflect 修改对象非导出字段的值

在 Go 的 `struct` 当中, 小写字段是非导出的, 即不可以从包外部访问.

*但非导出的字段在外部也并不是没有办法访问, 也不是不可以修改的*.

### reflect 取地址访问和修改非导出字段

函数 `reflect.NewAt`:

```cgo
// NewAt返回一个Value, 该指针表示一个指向指定类型, 使用p作为该指针.
func NewAt(typ Type, p unsafe.Pointer) Value {
	fl := flag(Ptr)
	t := typ.(*rtype)
	return Value{t.ptrTo(), p, fl}
}
```

`NewAt` 的实质上是构建了一个 `Ptr` 类型 `Value`, 该 `Value` 具有 `CanInterface()` 的特性.

经过 `Elem()` 之后, 又会增加 `CanSet()`, `CanAddr()` 两个特性.

> 注: NewAt 返回的"指针类型"的 reflect.Value. 指针的指向正是传入的地址 p.
> 要想获取"值(结构体)类型"的 reflect.Value, 只需要对返回值再执行 Elem() 操作即可.


函数 `UnsafeAddr()`

```cgo
// UnsafeAddr 返回指向 v 数据的指针.
// 适用于高级操作, 这些操作需要导入了 "unsafe" 包.
// 如果v不可寻址, 它会 panic
func (v Value) UnsafeAddr() uintptr {
	// TODO: deprecate
	if v.typ == nil {
		panic(&ValueError{"reflect.Value.UnsafeAddr", Invalid})
	}
	if v.flag&flagAddr == 0 {
		panic("reflect.Value.UnsafeAddr of unaddressable value")
	}
	return uintptr(v.ptr)
}
```

> 注意: 调用者必须是指针类型.

函数 `Set`:

```cgo
// Set 将 x 赋给值 v.
// 如果 CanSet 返回false, 则会 panic
// 和Go一样, x的值必须可分配给 v 的类型. (两者类型一致)
func (v Value) Set(x Value) {
	v.mustBeAssignable()
	x.mustBeExported() // do not let unexported x leak
	var target unsafe.Pointer
	if v.kind() == Interface {
		target = v.ptr
	}
	x = x.assignTo("reflect.Set", v.typ, target)
	if x.flag&flagIndir != 0 {
		typedmemmove(v.typ, v.ptr, x.ptr)
	} else {
		*(*unsafe.Pointer)(v.ptr) = x.ptr
	}
}
```

> 注意: Set() 强调两件事, 一个调用者本身可以被设置, 另外一个设置的值和调用者本身一致.


结构体类型定义:

```cgo
package t

type T struct {
    a string
}
```

**访问非导出字段**:

```cgo
func GetPtrUnExportFiled(s interface{}, filed string) reflec.Value {
    v := reflect.ValueOf(s).Elem().FiledByName(filed)
    // 必须要调用 Elem()
    return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}
```


**修改非导出字段**:

```cgo
func SetPtrUnExportFiled(s interface{}, filed string, val interface{}) error {
     v := GetPtrUnExportFiled(s, filed)
     rv := reflect.ValueOf(val)
     if v.Kind() != v.Kind() {
        return fmt.Errorf("invalid kind, expected kind: %v, got kind:%v", v.Kind(), rv.Kind())
     }
     
     v.Set(rv)
     return nil
}
```


### reflect 非取地址访问和修改非导出字段

函数 `New`:

```cgo
// New返回一个Value, 该值表示指向指定类型的新零值的指针. 也就是说, 返回的值的类型为PtrTo(typ).
func New(typ Type) Value {
	if typ == nil {
		panic("reflect: New(nil)")
	}
	t := typ.(*rtype)
	ptr := unsafe_New(t)
	fl := flag(Ptr)
	return Value{t.ptrTo(), ptr, fl}
}
```

`New` 和 `NewAt` 一样, 都是构建了一个 `Ptr` 类型 `Value`. 但是区别在于, `NewAt` 的指针是外部的, 而 `New` 的指针
是新创建的.

**访问非导出字段**:

```cgo
func GetUnExportFiled(s interface{}, filed string) (accessableField, addressableSource reflec.Value) {
    v := reflect.ValueOf(s)
    
    // 创建一个指向新地址的 addressableSource, 由于这个原因, 无法修改 s 本身的字段的值
    addressableSource = reflect.New(v.Type()).Elem()
    addressableSource.Set(s)
    
    // 使用指针的方式一样
    accessableField = addressableSource.FiledByName(filed)
    accessableField = reflect.NewAt(accessableField.Type(), unsafe.Pointer(accessableField.UnsafeAddr())).Elem()
    
    return accessableField, addressableSource
}
```

**设置非导出字段**:

```cgo
func SetUnExportFiled(s interface{}, filed string, val interface{}) error {
    v, addressableSource := GetUnExportFiled(s, filed)
    rv := reflect.ValueOf(val)
    if v.Kind() != v.Kind() {
        return fmt.Errorf("invalid kind, expected kind: %v, got kind:%v", v.Kind(), rv.Kind())
    }
         
    addressableSource.Set(rv)
    return nil
}
```

