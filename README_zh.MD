
- [English](./README_en.MD)

# KpSql _--(使您的数据库操作变的更加简易)_

## 安装

1. 在您的命令行中输入`go get github.com/KpnmServer/go-kpsql`并回车以安装该库到您的**golang环境**
2. 在您的*go源文件*顶部使用`import kpsql "github.com/KpnmServer/go-kpsql"`导入该库

## 使用

### 1. 创建`SqlDatabase`对象

您可以使用`kpsql.Open("[数据库驱动名称]", "[数据库连接地址]")`创建`SqlDatabase`对象  
该方法参数列表与`database/sql`库中的`Open`方法完全相同

例如
```golang
sqldb, err := kpsql.Open("mysql", 
	"user:password@tcp(127.0.0.1:3306)/mydatabase?charset=utf8&parseTime=true")
if err != nil {
	panic(err)
}
```

我们会在`Open`函数中调用`sql.DB.Ping`自动为您检测数据库是否连通, 如果不连通 _(即ping方法返回了err)_ 我们**仍会**将`SqlDatabase`与`error`一起返回  
您可以调用`SqlDatabase.DB()`方法获取原生`sql.DB`对象, 并重新执行Ping方法或进行其他操作

### 2. 创建数据库字段类型

数据库字段应为**struct**类型, 且*无需实现任何接口*, 但是需要**添加指定tag**到您的**公开字段**中

例如
```golang
type People struct{
	Name string `sql:"name"`
	Male bool `sql:"male"`
	Id int `sql:"id"`
}
```
此时 该类型的所有字段将在创建`SqlType`时被保存, 对应的键分别为`name`, `male`和`id`

**需要注意**, 非公开字段**不会参与**序列化与反序列化, **即使为其标注了tag**

例如
```golang
type People struct{
	name string `sql:"name"`
	Male bool `sql:"male"`
	Id int `sql:"id"`
}
```
此时 name字段**不会被**序列化与反序列化, 其他字段照常  
但是, 非公开类型**仍然可以参与**序列化与反序列化, 类型名称**不会影响到**我们的工作

如果您**未给字段分配**指定tag, 该字段也**不会参与**序列化与反序列化

### 3. 创建`SqlTable`对象

您有两种方式创建`SqlTable`对象, 第一种是使用`SqlDatabase.GetTable("[数据表名称]", [类型实例])`  
第二种是使用`SqlDatabase.GetTableBySqltype("[数据表名称]", [SqlType])`创建  
我们推荐您使用*第一种方法*创建数据表对象, 因为这样更简洁

例如
```golang
sqltb := sqldb.GetTable("myusers", &People{})
```
使用第二种方法
```golang
sqltype := kpsql.NewSqlType(&People{})
sqltb := sqldb.GetTableBySqltype("myusers", sqltype)
// 或者写成1行
sqltb := sqldb.GetTableBySqltype("myusers", kpsql.NewSqlType(&People{}))
```

如果您需要批量创建数据表, 您应首选第二种方法, 这样会减少创建不必要的对象  
例如
```golang
sqltype := kpsql.NewSqlType(&People{})
sqlgroups := make([]kpsql.SqlTable, 0, 100)
for i := 0; i < 100 ;i++ {
	sqlgroups = append(sqlgroups, sqldb.GetTableBySqltype(fmt.Sprintf("group_%d", i + 1), sqltype))
}
// sqlgroups[...]...
```

### 4. 使用`SqlTable`

接口定义如下
```golang
type SqlTable interface{
	Create()(err error) // 创建数据表
	Drop()(err error) // 删除数据表
	SqlType()(*SqlType) // 获取创建SqlTable实例时传入的SqlType对象
	Begin()(error) // 创建事务
	Rollback()(error) // 回滚事物 (幂等)
	Commit()(error) // 提交事物 (幂等)
	Insert(ins interface{})(n int64, err error) // 插入数据
	Delete(options ...sqloption)(n int64, err error) // 删除数据
	Update(ins interface{}, options ...sqloption)(n int64, err error) // 更新数据
	Select(options ...sqloption)(rows []interface{}, err error) // 查询数据
	SelectPrimary(ins interface{})(item interface{}, err error) // 根据主键自动生成where子句并查询1条数据
	Count(options ...sqloption)(n int64, err error) // 计数
}
```

`sqloption`为内部方法类型, 您不可以自己实现它, 我们提供了API让您可以生成`sqloption`  
API定义如下
```golang
type optionData struct{
	where WhereMap
	taglist []string
	limit []uint
	order string
}

type sqloption func(data *optionData)

func OptWhere(where WhereMap)(sqloption){ // 根据WhereMap生成选项
	return func(data *optionData){ data.where = where }
}

func OptWMap(arr ...interface{})(sqloption){ // 对MakeWMap方法的包装
	where := MakeWMap(arr...)
	return // 同上
}

func OptWMapAnd(arr ...interface{})(sqloption){ // 对MakeWMapAnd方法的包装
	where := MakeWMapAnd(arr...)
	return // 同上
}

func OptWMapOr(arr ...interface{})(sqloption){ // 对MakeWMapOr方法的包装
	where := MakeWMapOr(arr...)
	return // 同上
}

func OptWMapEq(arr ...interface{})(sqloption){ // 对MakeWMapEq方法的包装
	where := MakeWMapEq(arr...)
	return // 同上
}

func OptWMapEqAnd(arr ...interface{})(sqloption){ // 对MakeWMapEqAnd方法的包装
	where := MakeWMapEqAnd(arr...)
	return // 同上
}

func OptWMapEqOr(arr ...interface{})(sqloption){ // 对MakeWMapEqOr方法的包装
	where := MakeWMapEqOr(arr...)
	return // 同上
}

func OptLimit(limit ...uint)(sqloption){
	return func(data *optionData){ data.limit = limit }
}

func OptTags(tags ...string)(sqloption){
	return func(data *optionData){ data.taglist = tags }
}

func OptOrder(arr ...interface{})(sqloption){
	order := ""
	// 详见源码...
	return func(data *optionData){ data.order = order }
}
```

`WhereMap`对应sql中的`WHERE`字句, 如果您传递的值是`nil`, 那么**不会生成**where子句*(即匹配所有行)*  
格式如下
```golang
type WhereValue struct{
	Key   string // 键
	Cond  string // 条件
	Value interface{} // 值
	Next  string // 与下一个条件的逻辑连接关键字, 最后一项的`Next`字段将会被忽略
}
type WhereMap []WhereValue
```
您可以使用`kpsql.WhereMap{{"key1", "=", "value1", "AND"}, {"key2", "=", 2, ""}}`直接创建`WhereMap`, 格式化后的`WHERE`子句是
```sql
WHERE `key1` = ? AND `key2` = ?
```
其中, `?`为mysql的占位符, 实际所对应的值会在执行命令时由mysql进行填充以防止**注入攻击**  
您也可以使用我们提供的生成器(如:`MakeWMapEqAnd`)创建`WhereMap`  
方法定义如下
```golang

func MakeWMap(arr ...interface{})(wmap WhereMap){
	if len(arr) % 4 != 0 || len(arr) % 4 != 3 {
		panic("len(arr) % 4 != 0 || len(arr) % 4 != 3")
	}
	if len(arr) % 4 == 3 {
		arr = append(arr, "")
	}
	leng := len(arr) / 4
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: arr[i + 3].(string),
		})
	}
	return
}

func MakeWMapAnd(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 {
		panic("len(arr) % 3 != 0")
	}
	leng := len(arr) / 3
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: "AND",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapOr(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 {
		panic("len(arr) % 3 != 0")
	}
	leng := len(arr) / 3
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: "OR",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEq(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 || len(arr) % 3 != 2 {
		panic("len(arr) % 3 != 0 || len(arr) % 3 != 2")
	}
	if len(arr) % 3 == 2 {
		arr = append(arr, "")
	}
	leng := len(arr) / 3
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: arr[i + 2].(string),
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEqAnd(arr ...interface{})(wmap WhereMap){
	if len(arr) % 2 != 0 {
		panic("len(arr) % 2 != 0")
	}
	leng := len(arr) / 2
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: "AND",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEqOr(arr ...interface{})(wmap WhereMap){
	if len(arr) % 2 != 0 {
		panic("len(arr) % 2 != 0")
	}
	leng := len(arr) / 2
	wmap := make(WhereMap, 0, leng)
	for i := 0; i < leng ;i++ {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: "OR",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}
```
特别的, 对于`Update`方法, 如果没有传入`WhereMap`, 且字段之一的`sql_primary`值为`true`, 那么会自动生成一个`WhereMap`以匹配主键

对于`Insert`方法, 返回的`n`代表操作后主键自增的值 _(如果存在)_  
对于`Delete`, `Update`方法, 返回的`n`代表受到影响的行数  
`Select`方法返回的`rows`是创建map时所使用的实例的克隆的切片, 您需要在获取切片元素之后进行类型断言

例如
```golang
type People struct{
	Name string `sql:"name"`
	Male bool `sql:"male"`
	Id int `sql:"id"`
}

sqldb, err := kpsql.Open("mysql", 
	"user:password@tcp(127.0.0.1:3306)/mydatabase?charset=utf8&parseTime=true")
if err != nil {
	panic(err)
}
sqltb := sqldb.GetTable("myusers", &People{})

rows, err := sqltb.Select()
if err != nil {
	panic(err)
}
for _, p := range rows {
	people := p.(People)
	fmt.Printf("name: %s, is male: %v, id: %d", people.Name, people.Male, people.Id)
}
```

对于`Update`方法, `OptTags()`选项传入需要更新的**数据库字段名列表**_(即使用指定tag定义的值)_, 而**不是**结构体字段名  
  如果您的`taglist`为`nil`或者长度为0, 那么将会更新所有已绑定tag且不是主键的字段

对于`Insert`,`Update`和`SelectPrimary`方法, `ins`为**对象实例**, 我们并没有判断您传入的实例类型是否与之前传入`GetTable`方法中的一致  
  但是请注意, **如果对象的结构不同**, 极有可能导致程序**崩溃**

在`Insert`, `Delete`, `Update`, `Search`等等方法中, 我们自动帮您执行了所有需要执行的清理操作, 您无需为此担心  
  但是当您使用完数据库后您**仍需调用**`SqlDatabase.Close() /*与下句作用相同*/`或`SqlDatabase.DB().Close()`方法断开数据库连接, 以防止连接溢出

### 5.动态创建数据表

您可以为字段设置`sqlword`标签以定义Sql字段类型, 当您调用`SqlTable.Create()`创建数据表的时候它们会起作用  
例如
```golang
type People struct{
	Name string `sql:"name" sqlword:"VARCHAR(64)"`
	Male bool `sql:"male" sqlword:"BOOLEAN"`
	Id int `sql:"id" sqlword:"INTEGER UNIQUE"`
}
sqltable := sqldb.GetTable("myusers", &People{})
```
当您调用`sqltable.Create()`时, 我们会生成如下*sql*语句并调用
```sql
CREATE TABLE `myusers` (
	`name` VARCHAR(64) NOT NULL,
	`mail` BOOLEAN NOT NULL,
	`id` INTEGER UNIQUE NOT NULL
)
```
_(注:我们为了阅读方便在此处格式化了sql语句, 实际生成时为单行sql语句)_  
我们会自动为您添加`NOT NULL`后缀, 您无需自己添加. 但是出于安全原因, 该后缀也不可删除

如果您想定义主键, 请勿直接在`sqlword`后追加`PRIMARY KEY`, 我们推荐您为该字段定义`sql_primary`标签, 并将值设为`true`  
例如
```golang
type People struct{
	Name string `sql:"name" sqlword:"VARCHAR(64)"`
	Male bool `sql:"male" sqlword:"BOOLEAN"`
	Id int `sql:"id" sqlword:"INTEGER" sql_primary:"true"`
}
```
我们会生成如下语句
```sql
CREATE TABLE `myusers` (
	`name` VARCHAR(64) NOT NULL,
	`mail` BOOLEAN NOT NULL,
	`id` INTEGER NOT NULL,
	PRIMARY KEY (`id`)
)
```
您也可以定义复合主键
```golang
type Page struct{
	Book string `sql:"book" sqlword:"VARCHAR(64)" sql_primary:"true"`
	Page int `sql:"page" sqlword:"INTEGER" sql_primary:"true"`
	Content bool `sql:"data" sqlword:"TEXT"`
}
bookstore := sqldb.GetTable("bookstore", &Page{})
```
```sql
CREATE TABLE `bookstore` (
	`book` VARCHAR(64) NOT NULL,
	`page` INTEGER NOT NULL,
	`data` TEXT NOT NULL,
	PRIMARY KEY (`book`, `page`)
)
```

### 6.事务管理

我们提供了`SqlTable.Begin()`, `SqlTable.RollBack()`和`SqlTable.Commit()`三个方法来帮助您管理数据库事务  
在`Insert`,`Delete`,`Update`和`Select`方法中, 我们会检测当前是否有**正在进行的事务**, 如果没有, 我们会*自动创建一个新事务*并在**操作完成**后*自动提交*  
您可以使用`SqlTable.Begin()`开启一个新事务, 并在`SqlTable.Commit()`前执行您想要同步的操作.  
在此期间, 如果有**任何一个**数据库操作**发生异常**, 我们会**自动调用**`SqlTable.RollBack()`进行事务回滚, 无需您手动操作  
但我们并**不监听panic**错误, 如果您认为您的操作会导致panic, 我们**建议您**使用`recover`方法截获错误并**手动调用**`RollBack`方法

