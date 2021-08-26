
- [English](./README_en.MD)

# KpSql *--(使您的数据库操作变的更加简易)*

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

我们会在`Open`函数中调用`sql.DB.Ping`自动为您检测数据库是否连通, 如果不连通*(即ping方法返回了err)*我们**仍会**将`SqlDatabase`与`error`一起返回  
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

您有两种方式创建`SqlTable`对象, 第一种是使用`SqlDatabase.GetTable("[数据表名称]", [指向类型实例的指针])`  
第二种是使用`SqlDatabase.GetTableBySqltype("[数据表名称]", [SqlType])`创建  
我们推荐您使用*第一种方法*创建数据表对象, 因为这样更简洁

例如
```golang
sqltb := sqldb.GetTable("myusers", &People{})
```
使用第二种方法
```golang
sqltype := NewSqlType(&People{})
sqltb := sqldb.GetTableBySqltype("myusers", sqltype)
// 或者写成1行
sqltb := sqldb.GetTableBySqltype("myusers", NewSqlType(&People{}))
```

### 4. 使用`SqlTable`

我们提供了4个方法, 分别对应 增/删/改/查
方法定义如下
```golang
type SqlTable interface{
	Insert(ins interface{})(n int64, err error)
	Delete(wheremap WhereMap, limit ...uint)(n int64, err error)
	Update(ins interface{}, wheremap WhereMap, taglist []string, limit ...uint)(n int64, err error)
	Select(wheremap WhereMap, limit ...uint)(rows []interface{}, err error)
}
```
`Delete`, `Update`, `Select`方法中的`limit`参数对应到sql子句`LIMIT`  
  您可以传入1个数字(作为最大操作行数), 传入2个数字(前一个数字为偏移, 后一个数字为偏移后的最大操作行数), 或者不传入任何数字(操作所有符合条件的行)  

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
您可以使用`kpsql.WhereMap{{"key1", "=", "value1", "AND"}, {"key2", "=", 2, ""}}`创建`WhereMap`, 格式化后的`WHERE`子句是
```sql
WHERE `key1` = ? AND `key2` = ?
```
其中, `?`为mysql的占位符, 实际所对应的值会在执行命令时由mysql进行填充以防止**注入攻击**  
特别的, 对于`Update`方法, 如果传入的`wheremap`为`nil`且tag`sqlk`被指定, 那么会自动生成一个`WhereMap`, 指定该键的值为传入对象的值

对于`Insert`方法, 返回的`n`代表操作后主键自增的值*(如果存在)*  
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

rows, err := sqltb.Select(nil)
if err != nil {
	panic(err)
}
for _, p := range rows {
	people := p.(People)
	fmt.Printf("name: %s, is male: %v, id: %d", people.Name, people.Male, people.Id)
}
```

对于`Update`方法, `taglist`参数为需要更新的**数据库字段名***(即使用指定tag定义的值)*, 而**不是**结构体字段名  
  如果您的`taglist`为`nil`或者长度为0, 那么将会更新所有已绑定tag的字段

对于`Insert`和`Update`方法, `ins`为**指向对象实例的指针**, 我们并没有判断您传入的实例类型是否与之前传入`GetTable`方法中的一致  
  但是请注意, **如果对象的结构不同**, 极有可能导致程序**崩溃**

在`Insert`, `Delete`, `Update`和`Search`方法中, 我们自动帮您执行了所有需要执行的清理操作, 您无需为此担心  
  但是当您使用完数据库后您**仍需调用**`SqlDatabase.Close()`或`SqlDatabase.DB().Close()`方法断开数据库连接, 以防止连接溢出
